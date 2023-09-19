# px-object-controller

## Building

1. make all
2. make container
3. make deploy

## Running 

1. Setup environment variables:

```
export DOCKER_USER=""
export DOCKER_PASSWORD=""
export S3_ADMIN_ACCESS_KEY_ID=""
export S3_ADMIN_SECRET_ACCESS_KEY=""
export PURE_FB_ADMIN_ACCESS_KEY_ID=""
export PURE_FB_ADMIN_SECRET_ACCESS_KEY=""
```

2. Created required secrets:

```
kubectl -n kube-system create secret docker-registry pwxbuild --docker-username=${DOCKER_USER} --docker-password=${DOCKER_PASSWORD}
kubectl -n kube-system create secret generic px-object-s3-admin-credentials --from-literal=access-key-id=${S3_ADMIN_ACCESS_KEY_ID} --from-literal=secret-access-key=${S3_ADMIN_SECRET_ACCESS_KEY}
kubectl -n kube-system create secret generic px-object-fb-admin-credentials --from-literal=access-key-id=${PURE_FB_ADMIN_ACCESS_KEY_ID} --from-literal=secret-access-key=${PURE_FB_ADMIN_SECRET_ACCESS_KEY}
```

3. Create deployment:

```
 kubecl apply -f deploy/
```

## Running integration tests

1. Setup environment variable for testing:

```
export DOCKER_USER=""
export DOCKER_PASSWORD=""
export S3_ADMIN_ACCESS_KEY_ID=""
export S3_ADMIN_SECRET_ACCESS_KEY=""
export PURE_FB_ADMIN_ACCESS_KEY_ID=""
export PURE_FB_ADMIN_SECRET_ACCESS_KEY=""
```

2. Run `make integration-test-suite`

## Scripts
Build, deploy, and delete your local pods:
```
./hack/dev-refresh.sh
```
