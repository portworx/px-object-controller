set -o errexit
set -o nounset
set -o pipefail
set -x

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODE_GENERATOR_BASE=${GOPATH}/src/k8s.io
CODE_GENERATOR_PATH=${CODE_GENERATOR_BASE}/code-generator

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
bash "${CODE_GENERATOR_PATH}"/generate-groups.sh "deepcopy,client,lister,informer" \
  github.com/portworx/px-object-controller/client github.com/portworx/px-object-controller/client/apis \
  objectservice:v1alpha1 \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt