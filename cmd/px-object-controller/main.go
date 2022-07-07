package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/libopenstorage/openstorage/api/server/sdk"
	"github.com/libopenstorage/openstorage/bucket"
	"github.com/libopenstorage/openstorage/bucket/drivers/fake"
	"github.com/libopenstorage/openstorage/bucket/drivers/purefb"
	"github.com/libopenstorage/openstorage/bucket/drivers/s3"
	"github.com/libopenstorage/openstorage/pkg/correlation"
	"github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/portworx/kvdb"
	"github.com/portworx/px-object-controller/pkg/controller"
	"github.com/portworx/px-object-controller/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/zoido/yag-config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	envKubeconfig                  = "KUBECONFIG"
	envLogLevel                    = "LOG_LEVEL"
	envNamespace                   = "NAMESPACE"
	envWorkerThreads               = "WORKER_THREADS"
	envEnableLeaderElection        = "ENABLE_LEADER_ELECTION"
	envLeaderElectionNamespace     = "ENABLE_LEADER_ELECTION_NAMESPACE"
	envLeaderElectionLeaseDuration = "ENABLE_LEADER_ELECTION_LEASE_DURATION"
	envLeaderElectionRenewDeadline = "ENABLE_LEADER_ELECTION_RENEW_DEADLINE"
	envLeaderElectionRetryPeriod   = "ENABLE_LEADER_ELECTION_RETRY_PERIOD"
	envSDKPort                     = "SDK_PORT"
	envRestPort                    = "REST_PORT"
	envBucketDriver                = "BUCKET_DRIVER"
	envResyncPeriod                = "RESYNC_PERIOD"
	envRetryIntervalStart          = "RETRY_INTERVAL_START"
	envRetryIntervalMax            = "RETRY_INTERVAL_MAX"
	envS3AdminAccessKeyID          = "S3_ADMIN_ACCESS_KEY_ID"
	envS3AdminSecretAccessKey      = "S3_ADMIN_SECRET_ACCESS_KEY"
	envPureFBAdminAccessKeyID      = "PURE_FB_ADMIN_ACCESS_KEY_ID"
	envPureFBAdminSecretAccessKey  = "PURE_FB_ADMIN_SECRET_ACCESS_KEY"
)

var (
	kubeconfig                  string
	controllerNamespace         = "kube-system"
	logLevel                    = "debug"
	workers                     = 4
	leaderElection              = true
	leaderElectionNamespace     string
	leaderElectionLeaseDuration = 15 * time.Second
	leaderElectionRenewDeadline = 10 * time.Second
	leaderElectionRetryPeriod   = 5 * time.Second
	sdkPort                     = "18020"
	restPort                    = "18021"
	resyncPeriod                = 15 * time.Minute
	retryIntervalStart          = 1 * time.Second
	retryIntervalMax            = 5 * time.Minute
	s3AccessKeyID               = ""
	s3SecretAccessKey           = ""
	pureFBAccessKeyID           = ""
	pureFBSecretAccessKey       = ""
)

func parseFlags() error {
	y := yag.New()

	y.String(&kubeconfig, envKubeconfig, "Absolute path to the kubeconfig file. Required only when running out of cluster.")
	y.String(&controllerNamespace, envNamespace, "The namespace where the controller is running. Defaults to kube-system")
	y.String(&logLevel, envLogLevel, "Log level to use. Defaults to debug.")
	y.Bool(&leaderElection, envEnableLeaderElection, "Enables leader election.")
	y.String(&leaderElectionNamespace, envLeaderElectionNamespace, "The namespace where the leader election resource exists. Defaults to the pod namespace if not set.")
	y.Duration(&leaderElectionLeaseDuration, envLeaderElectionLeaseDuration, "Duration, in seconds, that non-leader candidates will wait to force acquire leadership. Defaults to 15 seconds.")
	y.Duration(&leaderElectionRenewDeadline, envLeaderElectionRenewDeadline, "Duration, in seconds, that the acting leader will retry refreshing leadership before giving up. Defaults to 10 seconds.")
	y.Duration(&leaderElectionRetryPeriod, envLeaderElectionRetryPeriod, "Duration, in seconds, the LeaderElector clients should wait between tries of actions. Defaults to 5 seconds.")

	y.Int(&workers, envWorkerThreads, "Number of worker threads.")
	y.String(&sdkPort, envSDKPort, "Openstorage SDK server port")
	y.String(&restPort, envRestPort, "Openstorage REST server port")
	y.Duration(&resyncPeriod, envResyncPeriod, "Resync interval of the controller.")
	y.Duration(&retryIntervalStart, envRetryIntervalStart, "Initial retry interval of failed bucket creation/access or deletion/revoke. It doubles with each failure, up to retry-interval-max. Default is 1 second.")
	y.Duration(&retryIntervalMax, envRetryIntervalMax, "Maximum retry interval of failed bucket/access creation or deletion/revoke. Default is 5 minutes.")
	y.String(&s3AccessKeyID, envS3AdminAccessKeyID, "Openstorage S3 Bucket Driver Access Key ID")
	y.String(&s3SecretAccessKey, envS3AdminSecretAccessKey, "Openstorage S3 Bucket Driver Access Secret Key")
	y.String(&pureFBAccessKeyID, envPureFBAdminAccessKeyID, "Openstorage Pure FB Bucket Driver Access Key ID")
	y.String(&pureFBSecretAccessKey, envPureFBAdminSecretAccessKey, "Openstorage Pure FB Bucket Driver Access Secret Key")

	return y.ParseEnv()
}

func main() {
	logrus.Infof("Staring PX controller version %v", version.Version)

	if err := parseFlags(); err != nil {
		logrus.Fatalf("failed to parse configuration variables. %v", err)
	}

	// Setting correlation logging
	correlation.RegisterGlobalHook()
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
	logrus.SetLevel(lvl)

	if err != nil {
		logrus.Fatalf("failed to initialize billing sink: %v", err)
	}

	// Create and start bucket drivers
	fakeBucketDriver := fake.New()
	driversMap := make(map[string]bucket.BucketDriver)
	driversMap[fakeBucketDriver.String()] = fakeBucketDriver
	go func() {
		if err := fakeBucketDriver.Start(); err != http.ErrServerClosed {
			logrus.Errorf("failed to start driver %s: %v", fakeBucketDriver.String(), err)
		}
	}()
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(s3AccessKeyID, s3SecretAccessKey, ""),
	}
	s3Driver, err := s3.New(s3Config)
	if err != nil {
		logrus.Fatalf("failed to create new s3 driver: %v", err)
	}
	driversMap[s3Driver.String()] = s3Driver
	pureFBConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(pureFBAccessKeyID, pureFBSecretAccessKey, ""),
	}
	pureFBConfig = pureFBConfig.WithDisableSSL(true).WithS3ForcePathStyle(true)
	pureFBDriver, err := purefb.New(pureFBConfig, pureFBAccessKeyID, pureFBSecretAccessKey)
	if err != nil {
		logrus.Fatalf("failed to create new s3 driver: %v", err)
	}
	driversMap[pureFBDriver.String()] = pureFBDriver

	// Create SDK object and start in background
	u, err := url.Parse("kv-mem://localhost")
	scheme := u.Scheme
	kv, err := kvdb.New(scheme, "openstorage", []string{u.String()}, nil, kvdb.LogFatalErrorCB)
	if err != nil {
		logrus.Fatalf("failed to initialize kvdb: %v", err)
	}
	if err := kvdb.SetInstance(kv); err != nil {
		logrus.Fatalf("failed set kvdb instance: %v", err)
	}
	sp, err := storagepolicy.Init()
	if err != nil {
		logrus.Fatalf("failed to initialize storage policy: %v", err)
	}
	sdkSocket := "/var/lib/osd/driver/sdk.sock"
	os.Remove(sdkSocket)
	if err := os.MkdirAll("/var/lib/osd/driver", 0750); err != nil {
		logrus.Fatalf("failed to initialize sdk socket location: %v", err)
	}
	sdkServer, err := sdk.New(&sdk.ServerConfig{
		Net:           "tcp",
		Address:       ":" + sdkPort,
		RestPort:      restPort,
		Socket:        sdkSocket,
		StoragePolicy: sp,
	})
	if err != nil {
		logrus.Fatalf("failed to start SDK server for driver: %v", err)
	}
	sdkServer.UseBucketDrivers(driversMap)
	go sdkServer.Start()

	// Create controller object
	ctrl, err := controller.New(&controller.Config{
		SdkUDS:             sdkSocket,
		BucketDrivers:      driversMap,
		RetryIntervalStart: retryIntervalStart,
		RetryIntervalMax:   retryIntervalMax,
	})
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	// Callback to start controller & sdk in goroutine
	run := func(context.Context) {
		// Run controller
		stopCh := make(chan struct{})
		go ctrl.Run(workers, stopCh)

		// Until SIGINT
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		close(stopCh)
	}

	// Start main loop with leader election
	if !leaderElection {
		logrus.Info("leader election not enabled")
		run(context.Background())
	} else {
		lockName := "px-object-controller-leader"
		// Create a new clientset for leader election to prevent throttling
		// due to px controller
		config, err := rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("failed to get in cluster config: %v", err)
		}
		leClientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			logrus.Fatalf("failed to create leaderelection client: %v", err)
		}
		le := leaderelection.NewLeaderElection(leClientset, lockName, run)

		if leaderElectionNamespace != "" {
			le.WithNamespace(leaderElectionNamespace)
		}
		le.WithLeaseDuration(leaderElectionLeaseDuration)
		le.WithRenewDeadline(leaderElectionRenewDeadline)
		le.WithRetryPeriod(leaderElectionRetryPeriod)
		if err := le.Run(); err != nil {
			logrus.Fatalf("failed to initialize leader election: %v", err)
		}
	}
}
