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
	"github.com/google/uuid"
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
		TestName: "[AWS S3] Controller test basic",
		TestConfig: specs.TestConfig{
			Namespace:   "default",
			Env:         basicEnv,
			BackendType: "S3Driver",
			Endpoint:    "s3.us-west-2.amazonaws.com",
			Region:      "us-west-2",
		},
		TestFunc: DynamicProvisionBasic,
	},
	{
		TestName: "[AWS S3] Basic bucket provision and access",
		TestConfig: specs.TestConfig{
			Namespace:    "default",
			Env:          basicEnv,
			BackendType:  "S3Driver",
			Endpoint:     "s3.us-west-2.amazonaws.com",
			Region:       "us-west-2",
			RetainBucket: true,
		},
		TestFunc: DynamicProvisionBasic,
	},
	{
		TestName: "[AWS S3] Basic bucket provision and access with retain",
		TestConfig: specs.TestConfig{
			Namespace:   "default",
			Env:         basicEnv,
			BackendType: "S3Driver",
			Endpoint:    "s3.us-west-1.amazonaws.com",
			Region:      "us-west-1",
		},
		TestFunc: DynamicProvisionBasic,
	},
	{
		TestName: "[AWS S3] Import existing bucket",
		TestConfig: specs.TestConfig{
			Namespace:    "default",
			Env:          basicEnv,
			BackendType:  "S3Driver",
			Endpoint:     "s3.us-west-1.amazonaws.com",
			Region:       "us-west-1",
			RetainBucket: true,
		},
		TestFunc: PreProvsionedBasic,
	},
	{
		TestName: "[Pure FB] Basic bucket provision and access",
		TestConfig: specs.TestConfig{
			Namespace:   "default",
			Env:         basicEnv,
			BackendType: "PureFBDriver",
			Endpoint:    "http://nfs.dogfood-skittles.dev.purestorage.com",
			Region:      "region-1",
		},
		TestFunc: DynamicProvisionBasic,
	},
	{
		TestName: "[Pure FB] Basic bucket provision and access with retain",
		TestConfig: specs.TestConfig{
			Namespace:    "default",
			Env:          basicEnv,
			BackendType:  "PureFBDriver",
			Endpoint:     "http://nfs.dogfood-skittles.dev.purestorage.com",
			Region:       "region-1",
			RetainBucket: true,
		},
		TestFunc: DynamicProvisionBasic,
	},
	{
		TestName: "[Pure FB] Import existing bucket",
		TestConfig: specs.TestConfig{
			Namespace:    "default",
			Env:          basicEnv,
			BackendType:  "PureFBDriver",
			Endpoint:     "http://nfs.dogfood-skittles.dev.purestorage.com",
			Region:       "region-1",
			RetainBucket: true,
		},
		TestFunc: PreProvsionedBasic,
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

func DynamicProvisionBasic(tc *types.TestCase) func(*testing.T) {
	return func(t *testing.T) {
		randID := uuid.NewString()[:8]
		className := fmt.Sprintf("pos-class-%s", randID)
		err := util.CreateBucketClass(objectClient, className, tc.TestConfig.RetainBucket, tc.TestConfig.Region, tc.TestConfig.BackendType, tc.TestConfig.Endpoint)
		if err != nil {
			t.Fatalf("failed to create object service class: %v", err)
		}

		claimName := fmt.Sprintf("pos-claim-%s", randID)
		err = util.CreateBucketClaim(objectClient, tc.TestConfig.Namespace, claimName, className)
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

			exists, err := CheckBucketExists(t, &tc.TestConfig, bucketClaim)
			if err != nil {
				t.Fatalf("check bucket exists failed: %v", err)
			}
			if !exists {
				t.Fatalf("check for bucket ID on backend system failed")
			}

			return nil
		})
		if err != nil {
			t.Fatalf("failed: %v", err)
		}

		accessName := fmt.Sprintf("pos-access-%s", randID)
		err = util.CreateBucketAccess(objectClient, tc.TestConfig.Namespace, accessName, className, claimName)
		if err != nil {
			t.Fatalf("failed to create bucketaccess")
		}

		var bucketAccessSecretName string
		err = RetryUntilSuccess("check if access granted", 5, 6, func() error {
			bucketAccess, err := objectClient.ObjectV1alpha1().PXBucketAccesses(tc.TestConfig.Namespace).Get(context.Background(), accessName, v1.GetOptions{})
			if err != nil {
				return err
			}

			if bucketAccess.Status == nil {
				return fmt.Errorf("bucketaccess %v access not granted yet", accessName)
			}

			if bucketAccess.Status.BackendType == "" || bucketAccess.Status.CredentialsSecretName == "" ||
				bucketAccess.Status.AccountId == "" || bucketAccess.Status.AccessGranted == false {
				return fmt.Errorf("bucket claim %s status invalid: %v", bucketAccess.Name, *bucketAccess.Status)
			}
			bucketAccessSecretName = bucketAccess.Status.CredentialsSecretName

			_, err = k8sClient.CoreV1().Secrets(tc.TestConfig.Namespace).Get(context.Background(), bucketAccess.Status.CredentialsSecretName, v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("bucket access granted but secret %s does not exist", bucketAccess.Status.CredentialsSecretName)
			}

			return nil
		})
		if err != nil {
			t.Fatalf("bucketaccess was never granted")
		}

		err = objectClient.ObjectV1alpha1().PXBucketAccesses(tc.TestConfig.Namespace).Delete(context.Background(), accessName, v1.DeleteOptions{})
		if err != nil {
			t.Fatalf("failed to delete bucketaccess %s", accessName)
		}
		err = RetryUntilSuccess("check if access revoked", 5, 6, func() error {
			_, err = k8sClient.CoreV1().Secrets(tc.TestConfig.Namespace).Get(context.Background(), bucketAccessSecretName, v1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				return nil
			}

			return fmt.Errorf("failed to check if secret %s still exists: %v", bucketAccessSecretName, err)
		})
		if err != nil {
			t.Fatalf("failed to check bucketaccess %s", accessName)
		}

		err = util.DeleteBucketClass(objectClient, className)
		if err != nil {
			t.Fatalf("delete object class failed: %v", err)
		}
		err = RetryUntilSuccess("check if class deleted", 5, 6, func() error {
			_, err := objectClient.ObjectV1alpha1().PXBucketClasses().Get(context.Background(), className, v1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				return nil
			}

			return fmt.Errorf("class %s still exists", className)
		})
		if err != nil {
			t.Fatalf("delete object class failed: %v", err)
		}

		err = util.DeleteBucketClaim(objectClient, tc.TestConfig.Namespace, claimName)
		if err != nil {
			t.Fatalf("delete object claim failed: %v", err)
		}
		err = RetryUntilSuccess("check if claim deleted", 5, 6, func() error {
			_, err := objectClient.ObjectV1alpha1().PXBucketClaims(tc.TestConfig.Namespace).Get(context.Background(), claimName, v1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				return nil
			}

			return fmt.Errorf("claim %s still exists", className)
		})
		if err != nil {
			t.Fatalf("delete object claim failed: %v", err)
		}

		exists, err := CheckBucketExists(t, &tc.TestConfig, bucketClaim)
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
		if exists {
			defer func() {
				CleanupBucket(t, &tc.TestConfig, bucketClaim)
			}()
		}
	}
}

func PreProvsionedBasic(tc *types.TestCase) func(*testing.T) {
	return func(t *testing.T) {
		randID := uuid.NewString()[:8]
		className := fmt.Sprintf("pos-class-%s", randID)
		err := util.CreateBucketClass(objectClient, className, tc.TestConfig.RetainBucket, tc.TestConfig.Region, tc.TestConfig.BackendType, tc.TestConfig.Endpoint)
		if err != nil {
			t.Fatalf("failed to create object service class: %v", err)
		}

		claimName := fmt.Sprintf("pos-claim-%s", randID)
		err = util.CreateBucketClaim(objectClient, tc.TestConfig.Namespace, claimName, className)
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

			exists, err := CheckBucketExists(t, &tc.TestConfig, bucketClaim)
			if err != nil {
				t.Fatalf("check bucket exists failed: %v", err)
			}
			if !exists {
				t.Fatalf("check for bucket ID on backend system failed")
			}

			return nil
		})
		if err != nil {
			t.Fatalf("failed: %v", err)
		}

		// Delete PBC and use bucket ID to import an existing bucket
		err = util.DeleteBucketClaim(objectClient, tc.TestConfig.Namespace, claimName)
		if err != nil {
			t.Fatalf("delete object claim failed: %v", err)
		}
		err = RetryUntilSuccess("check if claim deleted", 5, 6, func() error {
			_, err := objectClient.ObjectV1alpha1().PXBucketClaims(tc.TestConfig.Namespace).Get(context.Background(), claimName, v1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				return nil
			}

			return fmt.Errorf("claim %s still exists", className)
		})
		if err != nil {
			t.Fatalf("delete object claim failed: %v", err)
		}

		// Import existing bucket
		accessName := fmt.Sprintf("pos-access-%s", randID)
		err = util.CreateImportedBucketAccess(objectClient, tc.TestConfig.Namespace, accessName, className, bucketClaim.Status.BucketID)
		if err != nil {
			t.Fatalf("failed to create bucketaccess")
		}

		var bucketAccessSecretName string
		err = RetryUntilSuccess("check if access granted", 5, 6, func() error {
			bucketAccess, err := objectClient.ObjectV1alpha1().PXBucketAccesses(tc.TestConfig.Namespace).Get(context.Background(), accessName, v1.GetOptions{})
			if err != nil {
				return err
			}

			if bucketAccess.Status == nil {
				return fmt.Errorf("bucketaccess %v access not granted yet", accessName)
			}

			if bucketAccess.Status.BackendType == "" || bucketAccess.Status.CredentialsSecretName == "" ||
				bucketAccess.Status.AccountId == "" || bucketAccess.Status.AccessGranted == false {
				return fmt.Errorf("bucket claim %s status invalid: %v", bucketAccess.Name, *bucketAccess.Status)
			}
			bucketAccessSecretName = bucketAccess.Status.CredentialsSecretName

			_, err = k8sClient.CoreV1().Secrets(tc.TestConfig.Namespace).Get(context.Background(), bucketAccess.Status.CredentialsSecretName, v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("bucket access granted but secret %s does not exist", bucketAccess.Status.CredentialsSecretName)
			}

			return nil
		})
		if err != nil {
			t.Fatalf("bucketaccess was never granted")
		}

		err = objectClient.ObjectV1alpha1().PXBucketAccesses(tc.TestConfig.Namespace).Delete(context.Background(), accessName, v1.DeleteOptions{})
		if err != nil {
			t.Fatalf("failed to delete bucketaccess %s", accessName)
		}
		err = RetryUntilSuccess("check if access revoked", 5, 6, func() error {
			_, err = k8sClient.CoreV1().Secrets(tc.TestConfig.Namespace).Get(context.Background(), bucketAccessSecretName, v1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				return nil
			}

			return fmt.Errorf("failed to check if secret %s still exists: %v", bucketAccessSecretName, err)
		})
		if err != nil {
			t.Fatalf("failed to check bucketaccess %s", accessName)
		}

		err = util.DeleteBucketClass(objectClient, className)
		if err != nil {
			t.Fatalf("delete object class failed: %v", err)
		}
		err = RetryUntilSuccess("check if class deleted", 5, 6, func() error {
			_, err := objectClient.ObjectV1alpha1().PXBucketClasses().Get(context.Background(), className, v1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				return nil
			}

			return fmt.Errorf("class %s still exists", className)
		})
		if err != nil {
			t.Fatalf("delete object class failed: %v", err)
		}

		exists, err := CheckBucketExists(t, &tc.TestConfig, bucketClaim)
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
		if exists {
			defer func() {
				CleanupBucket(t, &tc.TestConfig, bucketClaim)
			}()
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

func getTestCaseS3Client(t *testing.T, tc *specs.TestConfig, endpoint, region string) (*s3.S3, error) {
	var s3Config *aws.Config
	switch tc.BackendType {
	case "S3Driver":
		s3Config = &aws.Config{
			Credentials: credentials.NewStaticCredentials(tc.Env.S3AdminAccessKeyID, tc.Env.S3AdminSecretAccessKey, ""),
		}
	case "PureFBDriver":
		s3Config = &aws.Config{
			Credentials: credentials.NewStaticCredentials(tc.Env.PureFBAdminAccessKeyID, tc.Env.PureFBAdminSecretAccessKey, ""),
		}
	}

	// Override the aws config with the region
	s3Config = s3Config.WithRegion(region)
	s3Config = s3Config.WithEndpoint(endpoint)

	sess, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}

	return s3.New(sess), nil
}

func CheckBucketExists(t *testing.T, tc *specs.TestConfig, pbc *v1alpha1.PXBucketClaim) (bool, error) {
	svc, err := getTestCaseS3Client(t, tc, pbc.Status.Endpoint, pbc.Status.Region)
	if err != nil {
		return false, err
	}

	resp, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil && strings.Contains(err.Error(), s3.ErrCodeNoSuchBucket) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	for _, b := range resp.Buckets {
		if *b.Name == pbc.Status.BucketID {
			return true, nil
		}
	}

	return false, nil
}

func CleanupBucket(t *testing.T, tc *specs.TestConfig, pbc *v1alpha1.PXBucketClaim) error {
	return CleanupBucketID(t, tc, pbc.Status.BucketID, pbc.Status.Endpoint, pbc.Status.Region)
}

func CleanupBucketID(t *testing.T, tc *specs.TestConfig, bucketID, endpoint, region string) error {
	svc, err := getTestCaseS3Client(t, tc, endpoint, region)
	if err != nil {
		return err
	}

	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketID),
	})
	if err != nil && strings.Contains(err.Error(), s3.ErrCodeNoSuchBucket) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}
