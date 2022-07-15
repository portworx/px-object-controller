//go:build integrationtest
// +build integrationtest

package integration_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/portworx/px-object-controller/client/apis/objectservice/v1alpha1"
	"github.com/portworx/px-object-controller/test/integration/specs"
	"github.com/portworx/px-object-controller/test/integration/types"
	"github.com/portworx/px-object-controller/test/integration/util"
	"github.com/sirupsen/logrus"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var testBasicCases = []types.TestCase{
	{
		TestName: "Controller test retain false",
		TestConfig: specs.TestConfig{
			Namespace: "kube-system",
			Env:       basicEnv,
		},
		TestFunc: BasicRun,
	},
	{
		TestName: "Controller test retain true",
		TestConfig: specs.TestConfig{
			Namespace:    "kube-system",
			RetainBucket: true,
			Env:          basicEnv,
		},
		TestFunc: BasicRun,
	},
}

var basicEnv = &specs.EnvConfig{
	S3AdminAccessKeyID:         os.Getenv("S3_ADMIN_ACCESS_KEY_ID"),
	S3AdminSecretAccessKey:     os.Getenv("S3_ADMIN_SECRET_ACCESS_KEY"),
	PureFBAdminAccessKeyID:     os.Getenv("PURE_FB_ADMIN_ACCESS_KEY_ID"),
	PureFBAdminSecretAccessKey: os.Getenv("PURE_FB_ADMIN_SECRET_ACCESS_KEY"),
}

func TestBasic(t *testing.T) {
	SetupTestEnvironment(t, basicEnv)
	for _, testCase := range testBasicCases {
		testCase.RunTest(t, k8sClient)
	}
}

func BasicRun(tc *types.TestCase) func(*testing.T) {
	return func(t *testing.T) {
		className := "pos-class-1"
		err := util.CreateObjectClass(objectClient, className, tc.TestConfig.RetainBucket, "us-west-2", "S3Driver", "s3.us-west-2.amazonaws.com")
		if err != nil {
			t.Fatalf("failed to create object service class: %v", err)
		}

		claimName := "pos-claim-1"
		err = util.CreateObjectClaim(objectClient, tc.TestConfig.Namespace, claimName, className)
		if err != nil {
			t.Fatalf("failed to create object service claim: %v", err)
		}

		var bucketClaim *v1alpha1.PXBucketClaim
		err = RetryUntilSuccess("get bucket claim", 5, 6, func() error {
			bucketClaim, err = objectClient.ObjectV1alpha1().PXBucketClaims(tc.TestConfig.Namespace).Get(context.Background(), claimName, v1.GetOptions{})
			if err != nil {
				return err
			}

			if bucketClaim.Status == nil {
				return fmt.Errorf("bucketclaim %v not provisioned yet", bucketClaim)
			}

			if bucketClaim.Status.BackendType == "" || bucketClaim.Status.BucketID == "" ||
				bucketClaim.Status.DeletionPolicy == "" || bucketClaim.Status.Provisioned == false || bucketClaim.Status.Region == "" {
				return fmt.Errorf("bucket claim %s status invalid: %v", bucketClaim.Name, *bucketClaim.Status)
			}

			desiredRetainPolicy := v1alpha1.PXBucketClaimDelete
			if tc.TestConfig.RetainBucket {
				desiredRetainPolicy = v1alpha1.PXBucketClaimRetain
			}
			if bucketClaim.Status.DeletionPolicy != desiredRetainPolicy {
				return fmt.Errorf("deletion policy for claim %v is invalid. expected: %v, got: %v", bucketClaim.Name, desiredRetainPolicy, bucketClaim.Status.DeletionPolicy)
			}

			exists, err := CheckBucketExists(t, &tc.TestConfig, bucketClaim.Status.BucketID)
			if err != nil {
				t.Fatalf("check bucket exists failed: %v", err)
			}
			if !exists {
				t.Fatalf("bucket provisioned but does not exist: %v", err)
			}

			return nil
		})
		if err != nil {
			t.Fatalf("failed: %v", err)
		}

		err = util.DeleteObjectClass(objectClient, className)
		if err != nil {
			t.Fatalf("delete object class failed: %v", err)
		}
		err = RetryUntilSuccess("check if class deleted", 3, 5, func() error {
			_, err := objectClient.ObjectV1alpha1().PXBucketClasses().Get(context.Background(), className, v1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				return nil
			}

			return fmt.Errorf("class %s still exists", className)
		})
		if err != nil {
			t.Fatalf("delete object class failed: %v", err)
		}

		err = util.DeleteObjectClaim(objectClient, tc.TestConfig.Namespace, claimName)
		if err != nil {
			t.Fatalf("delete object claim failed: %v", err)
		}
		err = RetryUntilSuccess("check if claim deleted", 3, 5, func() error {
			_, err := objectClient.ObjectV1alpha1().PXBucketClaims(tc.TestConfig.Namespace).Get(context.Background(), claimName, v1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				return nil
			}

			return fmt.Errorf("claim %s still exists", className)
		})
		if err != nil {
			t.Fatalf("delete object claim failed: %v", err)
		}

		exists, err := CheckBucketExists(t, &tc.TestConfig, bucketClaim.Status.BucketID)
		if err != nil {
			t.Fatalf("check bucket exists failed: %v", err)
		}
		if tc.TestConfig.RetainBucket {
			if !exists {
				t.Fatalf("retain bucket set but bucket does not exist")
			}
		} else {
			if exists {
				t.Fatalf("delete bucket set but bucket exists")
			}
		}
	}
}

func RetryUntilSuccess(name string, intervalSeconds int, maxBackoffs int, T func() error) error {
	for i := 0; i < maxBackoffs; i++ {
		err := T()
		if err == nil {
			return nil
		}

		logrus.Infof("%s failed, retrying again in %v seconds: %v", name, intervalSeconds, err)
		time.Sleep(time.Duration(intervalSeconds) * time.Second)
	}

	return fmt.Errorf("failed to %s after %v seconds", name, intervalSeconds*maxBackoffs)
}

func SetupTestEnvironment(t *testing.T, cfg *specs.EnvConfig) {
	// Setup PX Object controller
	err := specs.CreatePXObjectControllerDeployment(k8sClient, cfg)
	if err != nil {
		t.Fatalf("failed to create px-object deployment: %v", err)
	}
}

func CheckBucketExists(t *testing.T, tc *specs.TestConfig, bucketID string) (bool, error) {
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(tc.Env.S3AdminAccessKeyID, tc.Env.S3AdminSecretAccessKey, ""),
	}

	// Override the aws config with the region
	s3Config = s3Config.WithRegion("us-west-2")

	sess, err := session.NewSession(s3Config)
	if err != nil {
		return false, err
	}

	svc := s3.New(sess)
	_, err = svc.GetBucketLocation(&s3.GetBucketLocationInput{
		Bucket: aws.String(bucketID),
	})
	if err != nil && strings.Contains(err.Error(), s3.ErrCodeNoSuchBucket) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
