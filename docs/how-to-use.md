# How to use

## Create and use a PX Bucket Claim with AWS S3

### Provisioning a new bucket:

1. Create a new file named `pxbucketclass.yaml`:

```
apiVersion: object.portworx.io/v1alpha1
kind: PXBucketClass
metadata:
  name: pbclass-s3
region: us-west-1
deletionPolicy: Delete
parameters:
  object.portworx.io/backend-type: S3Driver
  object.portworx.io/endpoint: s3.us-west-1.amazonaws.com
```

2. Edit the above file with your desired AWS S3 region and endpoint

3. Create the PXBucketClass object:

```
kubectl apply -f pxbucketclass.yaml
```

4. Create a new file named `pxbucketclaim.yaml`:

```
apiVersion: object.portworx.io/v1alpha1
kind: PXBucketClaim
metadata:
  name: s3-pbc
  namespace: default
spec:
  bucketClassName: pbclass-s3
```

5. Create the PXBucketClaim object:

```
kubectl apply -f pxbucketclaim.yaml
```

6. Once the bucket is provisioned, it will be listed as Provisioned = true in the CustomResource:

```
$ kubectl get pxbucketclaim
NAME     PROVISIONED   BUCKETID                                     BACKENDTYPE
s3-pbc   true          px-os-06663fb0-d1bb-4b8a-914c-ac6595c2b721   S3Driver
```

### Providing Access to the PXBucketClaim:

1. Create a new file named `pxbucketaccess.yaml`:

```
apiVersion: object.portworx.io/v1alpha1
kind: PXBucketAccess
metadata:
  name: s3-pba
  namespace: default
spec:
  bucketClassName: pbclass-s3
  bucketClaimName: s3-pbc
```

2. Once the bucket access is granted, it will be marked as Access Granted = true in the CustomResource:

```
$ kubectl get pxbucketaccess
NAME     ACCESSGRANTED   CREDENTIALSSECRETNAME      BUCKETID                                     BACKENDTYPE
s3-pba   true            px-os-credentials-s3-pba   px-os-06663fb0-d1bb-4b8a-914c-ac6595c2b721   S3Driver
```

Additionally, a secret `px-os-credentials-s3-pba` will be created with all nessesary bucket info:

```
$ k get secret px-os-credentials-s3-pba -o yaml
apiVersion: v1
data:
  access-key-id: <ACCESS-KEY-ID>
  bucket-id: <BUCKET-ID>
  endpoint: <ENDPOINT>
  region: <REGION>
  secret-access-key: <SECRET-ACCESS-KEY>
kind: Secret
metadata:
  creationTimestamp: "2022-08-03T21:27:25Z"
  finalizers:
  - finalizers.object.portworx.io/access-secret
  name: px-os-credentials-s3-pba
  namespace: default
  resourceVersion: "16022682"
  uid: 49aecbbd-c911-48cf-95ea-9e9d30aba97c
type: Opaque
```

### Utilizing a PXBucketAccess credentials in an application:

1. In your application `deployment.yaml`, add all of the environment variables for your bucket(s) as Kubernetes deployment secret references:

```
    env:
    - name: S3_ACCESS_KEY
    valueFrom:
        secretKeyRef:
            name: px-os-credentials-s3-pba
            key: access-key-id
    - name: S3_SECRET_KEY
    valueFrom:
        secretKeyRef:
            name: px-os-credentials-s3-pba
            key: secret-access-key
    - name: S3_BUCKET_NAME
    valueFrom:
        secretKeyRef:
            name: px-os-credentials-s3-pba
            key: bucket-id
    - name: S3_ENDPOINT
    valueFrom:
        secretKeyRef:
            name: px-os-credentials-s3-pba
            key: endpoint
    - name: S3_REGION
    valueFrom:
        secretKeyRef:
            name: px-os-credentials-s3-pba
            key: region
```

2. Apply the updates to your `deployment.yaml`:

```
kubectl apply -f deployment.yaml
```

## Create and use a PX Bucket Claim with Pure Flashblade

NOTE - Same as above, but for Pure Flashblade