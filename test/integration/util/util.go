package util

import (
	"context"

	"github.com/portworx/px-object-controller/client/apis/objectservice/v1alpha1"
	clientset "github.com/portworx/px-object-controller/client/clientset/versioned"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateBucketClass creates an bucket class
func CreateBucketClass(objectClient *clientset.Clientset, name string, retainBucket bool, region string, backendType string, endpoint string) error {
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

// DeleteBucketClass deletes an bucket class
func DeleteBucketClass(objectClient *clientset.Clientset, name string) error {
	err := objectClient.ObjectV1alpha1().PXBucketClasses().Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// CreateBucketClaim creates an bucket claim
func CreateBucketClaim(objectClient *clientset.Clientset, namespace string, name string, bucketClassName string) error {
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

// DeleteBucketClaim deletes an bucket claim
func DeleteBucketClaim(objectClient *clientset.Clientset, namespace, name string) error {
	err := objectClient.ObjectV1alpha1().PXBucketClaims(namespace).Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// CreateBucketAccess creates a bucket claim
func CreateBucketAccess(objectClient *clientset.Clientset, namespace string, name string, bucketClassName, bucketClaimName string) error {
	_, err := objectClient.ObjectV1alpha1().PXBucketAccesses(namespace).Create(context.Background(), &v1alpha1.PXBucketAccess{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.BucketAccessSpec{
			BucketClassName: bucketClassName,
			BucketClaimName: bucketClaimName,
		},
	}, v1.CreateOptions{})
	return err
}

// CreateImportedBucketAccess creates a bucket claim
func CreateImportedBucketAccess(objectClient *clientset.Clientset, namespace string, name string, bucketClassName, existingBucketID string) error {
	_, err := objectClient.ObjectV1alpha1().PXBucketAccesses(namespace).Create(context.Background(), &v1alpha1.PXBucketAccess{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.BucketAccessSpec{
			BucketClassName:  bucketClassName,
			ExistingBucketId: existingBucketID,
		},
	}, v1.CreateOptions{})
	return err
}

// DeleteBucketAccess deletes a bucket claim
func DeleteBucketAccess(objectClient *clientset.Clientset, namespace string, name string) error {
	return objectClient.ObjectV1alpha1().PXBucketAccesses(namespace).Delete(context.Background(), name, v1.DeleteOptions{})
}
