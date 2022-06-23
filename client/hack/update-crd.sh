set -o errexit
set -o nounset
set -o pipefail
set -x

SCRIPT_ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)

# find or download controller-gen
CONTROLLER_GEN=$(which controller-gen)

$CONTROLLER_GEN crd:crdVersions=v1 paths=${SCRIPT_ROOT}/apis/objectservice/v1alpha1 output:dir=${SCRIPT_ROOT}/config/crd