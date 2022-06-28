package controller

import (
	"context"
	"fmt"

	"github.com/libopenstorage/openstorage/api"
	crdv1alpha1 "github.com/portworx/px-object-controller/client/apis/objectservice/v1alpha1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/portworx/px-object-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (ctrl *Controller) deleteBucket(ctx context.Context, pbc *crdv1alpha1.PXBucketClaim) {

	if pbc.Status == nil || !pbc.Status.Provisioned {
		logrus.Infof("bucket not yet provisioned. skipping backened delete")
		ctrl.bucketStore.Delete(pbc)
		return
	}

	// Issue delete if provisioned and deletionPolicy is delete
	if pbc.Status.DeletionPolicy == crdv1alpha1.PXBucketClaimRetain {
		logrus.Infof("skipping delete bucket as deletionPolicy was retain")
		ctrl.bucketStore.Delete(pbc)
		return
	}

	// Provisioned and deletionPolicy is delte. Delete the bucket here.
	_, err := ctrl.bucketClient.DeleteBucket(context.Background(), &api.BucketDeleteRequest{
		BucketId: pbc.Status.BucketID,
		Region:   pbc.Status.Region,
	})
	if err != nil {
		logrus.Infof("delete bucket %s failed: %v", pbc.Name, err)
	}
	ctrl.bucketStore.Delete(pbc)

	logrus.Infof("bucket %q deleted", pbc.Name)
}

func (ctrl *Controller) createBucket(ctx context.Context, pbc *crdv1alpha1.PXBucketClaim, pbclass *crdv1alpha1.PXBucketClass) error {
	_, err := ctrl.bucketClient.CreateBucket(ctx, &api.BucketCreateRequest{
		Name:   string(pbc.UID),
		Region: pbclass.Region,
	})
	if err != nil {
		logrus.Infof("create bucket %s failed: %v", pbc.Name, err)
	}

	logrus.Infof("bucket %q created", pbc.Name)
	if pbc.Status == nil {
		pbc.Status = &crdv1alpha1.BucketClaimStatus{}
	}
	pbc.Status.Provisioned = true
	pbc.Status.Region = pbclass.Region
	pbc.Status.DeletionPolicy = pbclass.DeletionPolicy
	pbc.Status.BucketID = string(pbc.UID)
	pbc, err = ctrl.k8sBucketClient.ObjectV1alpha1().PXBucketClaims(pbc.Namespace).Update(ctx, pbc, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	_, err = ctrl.storeBucketUpdate(pbc)
	return err
}

func getAccountName(pbclass *crdv1alpha1.PXBucketClass) string {
	return fmt.Sprintf("account-%v", pbclass.ObjectMeta.UID)
}

func getCredentialsSecretName(pba *crdv1alpha1.PXBucketAccess) string {
	if pba.Status != nil && pba.Status.CredentialsSecretName != "" {
		return pba.Status.CredentialsSecretName
	}
	return fmt.Sprintf("poc-credentials-%s", pba.Name)
}

func (ctrl *Controller) createAccess(ctx context.Context, pba *crdv1alpha1.PXBucketAccess, pbclass *crdv1alpha1.PXBucketClass, bucketID string) error {
	resp, err := ctrl.bucketClient.AccessBucket(ctx, &api.BucketGrantAccessRequest{
		BucketId:    bucketID,
		AccountName: getAccountName(pbclass),
	})
	if err != nil {
		logrus.Infof("create bucket access %s failed: %v", pba.Name, err)
		return err
	}

	accessData := make(map[string]string)
	accessData["accessKeyID"] = resp.Credentials.GetAccessKeyId()
	accessData["secretAccessKey"] = resp.Credentials.GetSecretAccessKey()

	// If secret exists, update it.
	credentialsSecretName := getCredentialsSecretName(pba)
	secret, err := ctrl.k8sClient.CoreV1().Secrets(pba.Namespace).Get(ctx, credentialsSecretName, metav1.GetOptions{})
	if k8s_errors.IsNotFound(err) {
		// Create if it doesn't exist
		secret, err = ctrl.k8sClient.CoreV1().Secrets(pba.Namespace).Create(
			ctx,
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      credentialsSecretName,
					Namespace: pba.Namespace,
				},
				StringData: accessData,
			},
			metav1.CreateOptions{},
		)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	logrus.Infof("bucket access %q created", pba.Name)
	if pba.Status == nil {
		pba.Status = &crdv1alpha1.BucketAccessStatus{}
	}
	pba.Status.AccessGranted = true
	pba.Status.CredentialsSecretName = secret.Name
	pba.Status.AccountId = resp.GetAccountId()
	pba.Status.BucketId = bucketID
	pba, err = ctrl.k8sBucketClient.ObjectV1alpha1().PXBucketAccesses(pba.Namespace).Update(ctx, pba, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	_, err = ctrl.storeAccessUpdate(pba)
	return err
}

func (ctrl *Controller) revokeAccess(ctx context.Context, pba *crdv1alpha1.PXBucketAccess) {

	if pba.Status == nil || !pba.Status.AccessGranted {
		logrus.Infof("bucket not yet provisioned. skipping backened delete")
		ctrl.accessStore.Delete(pba)
		return
	}

	// Provisioned and deletionPolicy is delte. Delete the bucket here.
	_, err := ctrl.bucketClient.RevokeBucket(context.Background(), &api.BucketRevokeAccessRequest{
		BucketId:  pba.Status.BucketId,
		AccountId: pba.Status.AccountId,
	})
	if err != nil {
		logrus.Infof("revoke bucket %s failed: %v", pba.Name, err)
	}

	err = ctrl.k8sClient.CoreV1().Secrets(pba.Namespace).Delete(ctx, pba.Status.CredentialsSecretName, metav1.DeleteOptions{})
	if k8s_errors.IsNotFound(err) {
		logrus.Infof("bucket access secret %s already deleted", pba.Status.CredentialsSecretName)
		return
	} else if err != nil {
		logrus.Infof("bucket access secret %s delete failed: %v", pba.Status.CredentialsSecretName, err)
		return
	}

	ctrl.accessStore.Delete(pba)
	logrus.Infof("bucket access %q deleted", pba.Name)
}

func (ctrl *Controller) storeBucketUpdate(bucket interface{}) (bool, error) {
	return utils.StoreObjectUpdate(ctrl.bucketStore, bucket, "bucket")
}

func (ctrl *Controller) storeAccessUpdate(access interface{}) (bool, error) {
	return utils.StoreObjectUpdate(ctrl.accessStore, access, "access")
}
