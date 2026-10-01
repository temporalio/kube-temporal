#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_DIR=${ROOT_DIR}/bin/code-generator
KUBE_CODEGEN_PATH=${CODEGEN_DIR}/kube_codegen.sh
SCRIPTS_DIR=${ROOT_DIR}/scripts
BOILERPLATE_PATH=${SCRIPTS_DIR}/boilerplate.go.txt

if [[ ! -f $KUBE_CODEGEN_PATH ]]; then
    echo "You must install Kubernetes code-generator before running kube-codegen-register.sh."
    echo "Hint: make gen-ensure"
    exit 1
fi

source "${KUBE_CODEGEN_PATH}"

kube::codegen::gen_register \
    --boilerplate ${BOILERPLATE_PATH} \
    "${ROOT_DIR}/api"
