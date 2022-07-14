package util

import (
	"context"

	"github.com/portworx/px-object-controller/client/apis/objectservice/v1alpha1"
	clientset "github.com/portworx/px-object-controller/client/clientset/versioned"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateObjectClass creates an object class
func CreateObjectClass(objectClient *clientset.Clientset, name string, retainBucket bool, region string, backendType string, endpoint string) error {
	deletionPolicy := v1alpha1.PXBucketClaimDelete
	if retainBucket {
		deletionPolicy = v1alpha1.PXBucketClaimRetain
	}

	_, err := objectClient.ObjectV1alpha1().PXBucketClasses().Create(context.Background(), &v1alpha1.PXBucketClass{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Region:         region,
		DeletionPolicy: deletionPolicy,
		Parameters: map[string]string{
			"object.portworx.io/backend-type": backendType,
			"object.portworx.io/endpoint":     endpoint,
		},
	}, v1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// DeleteObjectClass deletes an object class
func DeleteObjectClass(objectClient *clientset.Clientset, name string) error {
	err := objectClient.ObjectV1alpha1().PXBucketClasses().Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// CreateObjectClaim creates an object claim
func CreateObjectClaim(objectClient *clientset.Clientset, namespace string, name string, bucketClassName string) error {
	_, err := objectClient.ObjectV1alpha1().PXBucketClaims(namespace).Create(context.Background(), &v1alpha1.PXBucketClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.BucketClaimSpec{
			BucketClassName: bucketClassName,
		},
	}, v1.CreateOptions{})
	return err
}

// DeleteObjectClaim deletes an object claim
func DeleteObjectClaim(objectClient *clientset.Clientset, namespace, name string) error {
	err := objectClient.ObjectV1alpha1().PXBucketClaims(namespace).Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
