//go:build integrationtest
// +build integrationtest

package integration_test

import (
	"flag"
	"os"
	"testing"

	"github.com/portworx/px-object-controller/test/integration/types"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sClient *kubernetes.Clientset
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		logrus.Errorf("Setup failed with error: %v", err)
		os.Exit(1)
	}
	exitCode := m.Run()
	types.TestReporterInstance().PrintTestResult()
	os.Exit(exitCode)
}

func setup() error {
	// Parse flags
	var logLevel string
	var kubeconfig string
	var err error
	flag.StringVar(&logLevel,
		"log-level",
		"debug",
		"Log level")
	flag.StringVar(&kubeconfig,
		"kubeconfig",
		"",
		"Absolute path to the kubeconfig file. Required only when running out of cluster.")
	flag.Parse()

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(kubeconfig)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
	k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("failed to create k8s client: %v", err)
	}

	// Set log level
	logrusLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(logrusLevel)
	logrus.SetOutput(os.Stdout)

	return nil
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
