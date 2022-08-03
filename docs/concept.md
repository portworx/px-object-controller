# Overview

The Portworx Object Service allows storage admins to provision object storage buckets with various backing providers in a Kubernetes native way. This is achieved with a set of CustomResourceDefinitions managed by Stork:

1. [PXBucketClass](link/to/classref)
2. [PXBucketClaim](link/to/claimref)
3. [PXBucketAccess](link/to/accessref)

PX-Enterprise Object Service currently supports the following backing storage systems:

1. AWS S3
2. Pure FlashBlade Object

