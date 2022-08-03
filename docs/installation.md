# Prerequisites
* PX-Enterprise 2.12.0 or later
* Stork 2.12.0 or later
* Access to AWS S3 or Pure Flashblade secret access key ID and secret access key
* Kubernetes 1.17 cluster or later

# Installation
The Portworx Object Service objects are managed by Stork and will interact with a target PX-Enterprise instance. In the target PX-Enterprise instance sits the PX Object Service SDK which allows for bucket creation/deletion and providing/revoking access to buckets.

Additionally, the end user must provide access to the backend bucket service via environment variables. The following steps will allow PX-Enteprise to create and provide access to buckets on behalf of the credentials provided:

1. Create a Kubernetes new secret with S3 compliant secret access key ID and secret access key:

```
kubectl 
```