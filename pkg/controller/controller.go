package controller

import (
	"time"

	"github.com/libopenstorage/openstorage/pkg/correlation"
	crdv1alpha1 "github.com/portworx/px-object-controller/client/apis/pxobjectservice/v1alpha1"
	clientset "github.com/portworx/px-object-controller/client/clientset/versioned"
	"github.com/portworx/px-object-controller/client/clientset/versioned/scheme"
	bucketscheme "github.com/portworx/px-object-controller/client/clientset/versioned/scheme"
	informers "github.com/portworx/px-object-controller/client/informers/externalversions"
	bucketlisters "github.com/portworx/px-object-controller/client/listers/pxobjectservice/v1alpha1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/portworx/px-object-controller/pkg/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/klog/v2"
)

const (
	componentNameController = correlation.Component("pkg/controller")
)

var (
	logrus = correlation.NewPackageLogger(componentNameController)
)

// Config represents a configuration for creating a controller server
type Config struct {
	SdkUDS             string
	ResyncPeriod       time.Duration
	RetryIntervalStart time.Duration
	RetryIntervalMax   time.Duration
}

// Controller represents a controller server
type Controller struct {
	config *Config

	k8sBucketClient clientset.Interface
	k8sClient       kubernetes.Interface
	bucketClient    *client.Client
	eventRecorder   record.EventRecorder

	bucketQueue        workqueue.RateLimitingInterface
	bucketLister       bucketlisters.PXBucketClaimLister
	bucketListerSynced cache.InformerSynced
	bucketStore        cache.Store
	bucketFactory      informers.SharedInformerFactory
}

// New returns a new controller server
func New(cfg *Config) (*Controller, error) {

	// Get Openstorage Bucket SDK Client
	sdkBucketClient := client.NewClient(client.Config{
		SdkUDS: cfg.SdkUDS,
	})

	// Get general k8s clients
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("failed to create leaderelection client: %v", err)
	}
	k8sBucketClient, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Create new controller
	ctrl := &Controller{
		config:          cfg,
		k8sBucketClient: k8sBucketClient,
		k8sClient:       k8sClient,
		bucketClient:    sdkBucketClient,
	}

	// Create factory and informers
	factory := informers.NewSharedInformerFactory(k8sBucketClient, cfg.ResyncPeriod)
	bucketInformer := factory.Pxobjectservice().V1alpha1().PXBucketClaims()
	bucketInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueueBucketWork(obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueBucketWork(newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueueBucketWork(obj) },
		},
		ctrl.config.ResyncPeriod,
	)

	// Assign bucket CR listers and informers
	bucketRateLimiter := workqueue.NewItemExponentialFailureRateLimiter(ctrl.config.RetryIntervalStart, ctrl.config.RetryIntervalMax)
	ctrl.bucketFactory = factory
	ctrl.bucketStore = cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
	ctrl.bucketLister = bucketInformer.Lister()
	ctrl.bucketListerSynced = bucketInformer.Informer().HasSynced
	ctrl.bucketQueue = workqueue.NewNamedRateLimitingQueue(bucketRateLimiter, "px-object-controller-bucket")
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events(v1.NamespaceAll)})
	ctrl.eventRecorder = broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "px-object-controller"})
	bucketscheme.AddToScheme(scheme.Scheme)

	return ctrl, nil
}

// Run starts the Px Object Service controller
func (ctrl *Controller) Run(workers int, stopCh chan struct{}) {
	ctrl.bucketFactory.Start(stopCh)

	for i := 0; i < workers; i++ {
		go wait.Until(ctrl.bucketWorker, 0, stopCh)
	}

	<-stopCh
}

// bucketWorker is the main worker for PXBucketClaims.
func (ctrl *Controller) bucketWorker() {
	keyObj, quit := ctrl.bucketQueue.Get()
	if quit {
		return
	}
	defer ctrl.bucketQueue.Done(keyObj)

	if err := ctrl.processBucket(keyObj.(string)); err != nil {
		// Rather than wait for a full resync, re-add the key to the
		// queue to be processed.
		ctrl.bucketQueue.AddRateLimited(keyObj)
		klog.V(4).Infof("Failed to sync bucket %q, will retry again: %v", keyObj.(string), err)
	} else {
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		ctrl.bucketQueue.Forget(keyObj)
	}
}

func (ctrl *Controller) processBucket(key string) error {
	klog.V(5).Infof("syncBucketClaimByKey[%s]", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	klog.V(5).Infof("processBucket: bucket namespace [%s] name [%s]", namespace, name)
	if err != nil {
		klog.Errorf("error getting namespace & name of bucketclaim %q to get bucketclaim from informer: %v", key, err)
		return nil
	}
	bucketClaim, err := ctrl.bucketLister.PXBucketClaims(namespace).Get(name)
	if err == nil {
		klog.V(5).Infof("Creating bucketclaim %q", key)
		return ctrl.createBucket(bucketClaim)
	}
	if err != nil && !errors.IsNotFound(err) {
		klog.V(2).Infof("error getting bucketclaim %q from informer: %v", key, err)
		return err
	}
	// The bucketclaim is not in informer cache, the event must have been "delete"
	vsObj, found, err := ctrl.bucketStore.GetByKey(key)
	if err != nil {
		klog.V(2).Infof("error getting bucketclaim %q from cache: %v", key, err)
		return nil
	}
	if !found {
		// The controller has already processed the delete event and
		// deleted the bucketclaim from its cache
		klog.V(2).Infof("deletion of bucketclaim %q was already processed", key)
		return nil
	}
	bucketclaim, ok := vsObj.(*crdv1alpha1.PXBucketClaim)
	if !ok {
		klog.Errorf("expected vs, got %+v", vsObj)
		return nil
	}

	klog.V(5).Infof("deleting bucketclaim %q", key)
	ctrl.deleteBucket(bucketclaim)

	return nil
}

// enqueueBucketClaimWork adds bucketclaim to given work queue.
func (ctrl *Controller) enqueueBucketWork(obj interface{}) {
	// Beware of "xxx deleted" events
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}
	if bucket, ok := obj.(*crdv1alpha1.PXBucketClaim); ok {
		objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(bucket)
		if err != nil {
			klog.Errorf("failed to get key from object: %v, %v", err, bucket)
			return
		}
		klog.V(5).Infof("enqueued %q for sync", objName)
		ctrl.bucketQueue.Add(objName)
	}
}
