//
// Copyright 2019 Insolar Technologies GmbH
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
//

package bootstrap

import (
	"github.com/insolar/insolar/bootstrap/rootdomain"
	"github.com/insolar/insolar/insolar"
)

var (
	// ContractRootDomain is the root domain contract reference.
	ContractRootDomain = rootdomain.RootDomain.Ref()
	// ContractNodeDomain is the node domain contract reference.
	ContractNodeDomain = rootdomain.GenesisRef(insolar.GenesisNameNodeDomain)
	// ContractNodeRecord is the node contract reference.
	ContractNodeRecord = rootdomain.GenesisRef(insolar.GenesisNameNodeRecord)
	// ContractRootMember is the root member contract reference.
	ContractRootMember = rootdomain.GenesisRef("root" + insolar.GenesisNameMember)
	// ContractRootWallet is the root wallet contract reference.
	ContractRootWallet = rootdomain.GenesisRef("root" + insolar.GenesisNameWallet)
	// ContractMDAdminMember is the md admin wallet contract reference.
	ContractMDAdminMember = rootdomain.GenesisRef("mdadmin" + insolar.GenesisNameMember)
	// ContractOracleMembers is the oracles members contract reference.
	ContractOracleMembers = map[string]insolar.Reference{}
	// ContractMDWallet is the md wallet contract reference.
	ContractMDWallet = rootdomain.GenesisRef("md" + insolar.GenesisNameWallet)
)
