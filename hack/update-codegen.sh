#!/usr/bin/env bash

# Copyright 2020 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

source $(dirname $0)/../vendor/knative.dev/hack/codegen-library.sh

# If we run with -mod=vendor here, then generate-groups.sh looks for vendor files in the wrong place.
export GOFLAGS=-mod=

echo "=== Update Codegen for $MODULE_NAME"

# NO NEED FOR THIS!
# Compute _example hash for all configmaps.
# group "Generating checksums for configmap _example keys"
#
# ${REPO_ROOT_DIR}/hack/update-checksums.sh

group "Kubernetes Codegen"

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  "github.com/aliok/websocket-channel/pkg/client" "github.com/aliok/websocket-channel/pkg/apis" \
  "channels:v1alpha1" \
  --go-header-file ${REPO_ROOT_DIR}/hack/boilerplate/boilerplate.go.txt

# DO NOT DO THE FOLLOWING! No default config available yet!
#
# Deep copy config
#${GOPATH}/bin/deepcopy-gen \
#  -O zz_generated.deepcopy \
#  --go-header-file ${REPO_ROOT_DIR}/hack/boilerplate/boilerplate.go.txt \
#  -i github.com/aliok/websocket-channel/pkg/apis/config \
#  -i github.com/aliok/websocket-channel/pkg/apis/messaging/config \

# DO NOT DO THE FOLLOWING! No duck types available yet!
#
## Only deepcopy the Duck types, as they are not real resources.
#${CODEGEN_PKG}/generate-groups.sh "deepcopy" \
#  knative.dev/eventing/pkg/client knative.dev/eventing/pkg/apis \
#  "duck:v1alpha1 duck:v1beta1 duck:v1" \
#  --go-header-file ${REPO_ROOT_DIR}/hack/boilerplate/boilerplate.go.txt

group "Knative Codegen"

## Knative Injection
${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
  "github.com/aliok/websocket-channel/pkg/client" "github.com/aliok/websocket-channel/pkg/apis" \
  "channels:v1alpha1" \
  --go-header-file ${REPO_ROOT_DIR}/hack/boilerplate/boilerplate.go.txt

group "Deepcopy Gen"

# Depends on generate-groups.sh to install bin/deepcopy-gen
${GOPATH}/bin/deepcopy-gen \
  -O zz_generated.deepcopy \
  --go-header-file ${REPO_ROOT_DIR}/hack/boilerplate/boilerplate.go.txt

group "Update deps post-codegen"

# Make sure our dependencies are up-to-date
${REPO_ROOT_DIR}/hack/update-deps.sh
