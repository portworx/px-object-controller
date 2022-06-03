package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/libopenstorage/openstorage/pkg/correlation"
	"github.com/portworx/px-object-controller/pkg/controller"
	"github.com/portworx/px-object-controller/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/zoido/yag-config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
)

var (
	kubeconfig                  string
	controllerNamespace         = "kube-system"
	logLevel                    = "debug"
	threads                     = 10
	usageInterval               = 5 * time.Second
	collectorSource             = "fake"
	leaderElection              = true
	leaderElectionNamespace     string
	leaderElectionLeaseDuration = 15 * time.Second
	leaderElectionRenewDeadline = 10 * time.Second
	leaderElectionRetryPeriod   = 5 * time.Second
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
	y.Int(&threads, envWorkerThreads, "Number of worker threads.")

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

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(kubeconfig)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("failed to create leaderelection client: %v", err)
	}
	_ = k8sClient

	if err != nil {
		logrus.Fatalf("failed to initialize billing sink: %v", err)
	}

	// Create controller object
	ctrl, err := controller.New(&controller.Config{})
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	// Callback to start controller in goroutine
	run := func(context.Context) {
		// run...
		stopCh := make(chan struct{})
		go ctrl.Run(threads, stopCh)

		// ...until SIGINT
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

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
