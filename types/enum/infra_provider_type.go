// Copyright 2023 Nxenv, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enum

type InfraProviderType string

func (InfraProviderType) Enum() []interface{} { return toInterfaceSlice(providerTypes) }

var providerTypes = []InfraProviderType{
	InfraProviderTypeDocker, InfraProviderTypeNxenvGCP, InfraProviderTypeNxenvCloud, InfraProviderTypeHybridVMGCP,
}

const (
	InfraProviderTypeDocker       InfraProviderType = "docker"
	InfraProviderTypeNxenvGCP   InfraProviderType = "nxenv_gcp"
	InfraProviderTypeNxenvCloud InfraProviderType = "nxenv_cloud"
	InfraProviderTypeHybridVMGCP  InfraProviderType = "hybrid_vm_gcp"
)
