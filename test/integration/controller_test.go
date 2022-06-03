//go:build integrationtest
// +build integrationtest

package integration_test

import (
	"testing"

	"github.com/portworx/px-object-controller/test/integration/specs"
	"github.com/portworx/px-object-controller/test/integration/types"
)

var testBasicCases = []types.TestCase{
	{
		TestName: "Controller test",
		TestConfig: specs.TestConfig{
			PdsUsername:     "username",
			PdsPassword:     "password",
			PdsClientID:     "clientid",
			PdsClientSecret: "clientsecret",
		},
		TestFunc: BasicRun,
	},
}

func TestBasic(t *testing.T) {
	for _, testCase := range testBasicCases {
		testCase.RunTest(t, k8sClient)
	}
}

func BasicRun(tc *types.TestCase) func(*testing.T) {
	return func(t *testing.T) {

	}
}
