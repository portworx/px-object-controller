package controller

import (
	"github.com/libopenstorage/openstorage/pkg/correlation"
	storagelisters "github.com/portworx/px-object-controller/client/listers/volumesnapshot/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/klog/v2"
)

const (
	componentNameController = correlation.Component("pkg/controller")
)

var (
	logger = correlation.NewPackageLogger(componentNameController)
)

// Config represents a configuration for creating a controller server
type Config struct {
}

// Server represents a controller server
type Controller struct {
	config *Config

	// clientset   clientset.Interface
	client      kubernetes.Interface
	objectQueue workqueue.RateLimitingInterface
	accessQueue workqueue.RateLimitingInterface

	bucketLister       storagelisters.VolumeSnapshotLister
	bucketListerSynced cache.InformerSynced
	accessLister       storagelisters.VolumeSnapshotContentLister
	accessListerSynced cache.InformerSynced
}

// New returns a new controller server
func New(cfg *Config) (*Controller, error) {
	return &Controller{
		config: cfg,
	}, nil
}

// Run starts the Px Object Service controller
func (ctrl *Controller) Run(workers int, stopCh chan struct{}) {
	for i := 0; i < workers; i++ {
		go wait.Until(ctrl.bucketWorker, 0, stopCh)
		go wait.Until(ctrl.accessWorker, 0, stopCh)
	}

	<-stopCh
}

// bucketWorker is the main worker for PXBucketClaims.
func (ctrl *Controller) bucketWorker() {
	keyObj, quit := ctrl.snapshotQueue.Get()
	if quit {
		return
	}
	defer ctrl.snapshotQueue.Done(keyObj)

	if err := ctrl.syncSnapshotByKey(keyObj.(string)); err != nil {
		// Rather than wait for a full resync, re-add the key to the
		// queue to be processed.
		ctrl.snapshotQueue.AddRateLimited(keyObj)
		klog.V(4).Infof("Failed to sync snapshot %q, will retry again: %v", keyObj.(string), err)
	} else {
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		ctrl.snapshotQueue.Forget(keyObj)
	}
}

// bucketWorker is the main worker for PXBucketClaims.
func (ctrl *Controller) accessWorker() {
	keyObj, quit := ctrl.snapshotQueue.Get()
	if quit {
		return
	}
	defer ctrl.snapshotQueue.Done(keyObj)

	if err := ctrl.syncSnapshotByKey(keyObj.(string)); err != nil {
		// Rather than wait for a full resync, re-add the key to the
		// queue to be processed.
		ctrl.snapshotQueue.AddRateLimited(keyObj)
		klog.V(4).Infof("Failed to sync snapshot %q, will retry again: %v", keyObj.(string), err)
	} else {
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		ctrl.snapshotQueue.Forget(keyObj)
	}
}
