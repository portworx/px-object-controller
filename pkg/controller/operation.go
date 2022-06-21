package controller

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	crdv1alpha1 "github.com/portworx/px-object-controller/client/apis/pxobjectservice/v1alpha1"
)

func (ctrl *Controller) deleteBucket(pbc *crdv1alpha1.PXBucketClaim) {
	_, err := ctrl.bucketClient.DeleteBucket(context.Background(), &api.BucketDeleteRequest{
		BucketId: pbc.Name,
	})
	if err != nil {
		logrus.Infof("delete bucket %s failed: %v", pbc.Name, err)
	}

	logrus.Infof("bucket %q deleted", pbc.Name)
}

func (ctrl *Controller) createBucket(pbc *crdv1alpha1.PXBucketClaim) error {
	_, err := ctrl.bucketClient.CreateBucket(context.Background(), &api.BucketCreateRequest{
		Name:   pbc.Name,
		Region: "uswest-2",
	})
	if err != nil {
		logrus.Infof("create bucket %s failed: %v", pbc.Name, err)
	}

	logrus.Infof("bucket %q created", pbc.Name)
	return nil
}