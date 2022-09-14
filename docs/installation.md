# Prerequisites

To install Portworx Object Service, you must meet the following prerequisites:

* Use PX-Enterprise 2.12.0 or newer
* Use Stork 2.12.0 or newer
* Provide access to an AWS S3 or Pure Flashblade secret access key ID and secret access key
* Use Kubernetes 1.17 cluster or newer

# Installation

Portworx Object Service objects are managed by Stork and will interact with a target Portworx Enterprise instance. The PX Object Service SDK is located in the target Portworx Enterprise instance. This allows for bucket creation, bucket deletion, and providing or revoking access to buckets.

Additionally, you must provide access to the backend bucket service through environment variables. Since the Portworx Object Service is in Alpha, extra steps are required to enable and set up the Portworx Object Service controller. The following steps will allow Portworx Enteprise to create and provide access to buckets on behalf of the credentials provided:

1. Enable the Portworx Object Service controller flag in Stork by adding the following `args` to your StorageCluster spec:

    ```
    spec:
      ...
      stork:
        enabled: true
        args:
          px-object-controller: true
    ```

2. Create a new Kubernetes secret with your S3-compliant secret access key ID and secret access key:

    ```
    kubectl create secret generic px-object-s3-admin-credentials \ 
        --from-literal=access-key-id=ACCESS_KEY --from-literal=secret-access-key=SECRET_ACCESS_KEY 
    ```

3. Add environment variables for bucket credentials to your StorageCluster spec:

    ```
    spec:
        env:
            OBJECT_SERVICE_S3_ACCESS_KEY_ID: <access-key-id>
            OBJECT_SERVICE_S3_SECRET_ACCESS_KEY: <secret-access-key>
    ```

  _or_

    ```
    spec:
        env:
            OBJECT_SERVICE_FB_ACCESS_KEY_ID: <access-key-id>
            OBJECT_SERVICE_FB_SECRET_ACCESS_KEY: <secret-access-key>
    ```
