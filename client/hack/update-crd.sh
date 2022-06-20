#set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)

# find or download controller-gen
CONTROLLER_GEN=$(which controller-gen)

# if [ "$CONTROLLER_GEN" = "" ]
# then
#   TMP_DIR=$(mktemp -d);
#   cd $TMP_DIR;
#   go mod init tmp;
#   go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0;
#   rm -rf $TMP_DIR;
#   CONTROLLER_GEN=$(which controller-gen)
# fi

# if [ "$CONTROLLER_GEN" = "" ]
# then
#   echo "ERROR: failed to get controller-gen";
#   exit 1;
# fi

$CONTROLLER_GEN crd:crdVersions=v1 paths=${SCRIPT_ROOT}/apis/pxobjectservice/v1alpha1 output:dir=${SCRIPT_ROOT}/config/crd