# Reference

## Environment Variables

### PX-Enterprise

* `OBJECT_SERVICE_S3_ACCESS_KEY_ID`: An AWS S3 Access Key ID credential generated in the AWS Portal.
* `OBJECT_SERVICE_S3_SECRET_ACCESS_KEY`: An AWS S3 Secret Access Key credential generated in the AWS Portal.
* `OBJECT_SERVICE_FB_ACCESS_KEY_ID`: A Pure FlashBlade Access Key ID credential provided by the FlashBlade admin.
* `OBJECT_SERVICE_FB_SECRET_ACCESS_KEY`: A Pure FlashBlade Secret Access Key credential provided by the FlashBlade admin.

### Stork

* `WORKER_THREADS`: The number of worker threads to use in the Portworx Object Service Stork controller
* `RETRY_INTERVAL_START`: Initial retry interval of failed bucket creation/access or deletion/revoke. It doubles with each failure, up to retry-interval-max. Default is 1 second.
* `RETRY_INTERVAL_MAX`: Maximum retry interval of failed bucket/access creation or deletion/revoke. Default is 5 minutes.

## CustomResourceDefinitions

### PXBucketClass

```
apiVersion: object.portworx.io/v1alpha1
kind: PXBucketClass
metadata:
  name: <NAME>
region: <REGION>
deletionPolicy: <Delete or Retain> - # Indicates whether or not to execute a deletion call to the backing storage solution on PXBucketClaim deletion.
parameters:
  object.portworx.io/backend-type: [ S3Driver | PureFBDriver ]
  object.portworx.io/endpoint: <S3_ENDPOINT>
```

### PXBucketClaim

```
apiVersion: object.portworx.io/v1alpha1
kind: PXBucketClaim
metadata:
  name: <NAME>
  namespace: <NAMESPACE>
spec:
  bucketClassName: <BUCKET_CLASS_NAME>
```

### PXBucketAccess

```
apiVersion: object.portworx.io/v1alpha1
kind: PXBucketAccess
metadata:
  name: <NAME>
  namespace: <NAMESPACE>
spec:
  bucketClassName: <BUCKET_CLASS_NAME>
  bucketClaimName: <BUCKET_CLAIM_NAME>
```