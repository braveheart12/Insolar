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

package rootdomain

import (
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/logicrunner/common"
	"github.com/insolar/insolar/logicrunner/goplugin/foundation"
)

// PrototypeReference to prototype of this contract
// error checking hides in generator
var PrototypeReference, _ = insolar.NewReferenceFromBase58("1111hjS6s4X4fxQLMXXi4ozdNCfSXSxNWsD3kvocTU.11111111111111111111111111111111")

// RootDomain holds proxy type
type RootDomain struct {
	Reference insolar.Reference
	Prototype insolar.Reference
	Code      insolar.Reference
}

// ContractConstructorHolder holds logic with object construction
type ContractConstructorHolder struct {
	constructorName string
	argsSerialized  []byte
}

// AsChild saves object as child
func (r *ContractConstructorHolder) AsChild(objRef insolar.Reference) (*RootDomain, error) {
	ref, err := common.CurrentProxyCtx.SaveAsChild(objRef, *PrototypeReference, r.constructorName, r.argsSerialized)
	if err != nil {
		return nil, err
	}
	return &RootDomain{Reference: ref}, nil
}

// AsDelegate saves object as delegate
func (r *ContractConstructorHolder) AsDelegate(objRef insolar.Reference) (*RootDomain, error) {
	ref, err := common.CurrentProxyCtx.SaveAsDelegate(objRef, *PrototypeReference, r.constructorName, r.argsSerialized)
	if err != nil {
		return nil, err
	}
	return &RootDomain{Reference: ref}, nil
}

// GetObject returns proxy object
func GetObject(ref insolar.Reference) (r *RootDomain) {
	return &RootDomain{Reference: ref}
}

// GetPrototype returns reference to the prototype
func GetPrototype() insolar.Reference {
	return *PrototypeReference
}

// GetImplementationFrom returns proxy to delegate of given type
func GetImplementationFrom(object insolar.Reference) (*RootDomain, error) {
	ref, err := common.CurrentProxyCtx.GetDelegate(object, *PrototypeReference)
	if err != nil {
		return nil, err
	}
	return GetObject(ref), nil
}

// NewRootDomain is constructor
func NewRootDomain() *ContractConstructorHolder {
	var args [0]interface{}

	var argsSerialized []byte
	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		panic(err)
	}

	return &ContractConstructorHolder{constructorName: "NewRootDomain", argsSerialized: argsSerialized}
}

// GetReference returns reference of the object
func (r *RootDomain) GetReference() insolar.Reference {
	return r.Reference
}

// GetPrototype returns reference to the code
func (r *RootDomain) GetPrototype() (insolar.Reference, error) {
	if r.Prototype.IsEmpty() {
		ret := [2]interface{}{}
		var ret0 insolar.Reference
		ret[0] = &ret0
		var ret1 *foundation.Error
		ret[1] = &ret1

		res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "GetPrototype", make([]byte, 0), *PrototypeReference)
		if err != nil {
			return ret0, err
		}

		err = common.CurrentProxyCtx.Deserialize(res, &ret)
		if err != nil {
			return ret0, err
		}

		if ret1 != nil {
			return ret0, ret1
		}

		r.Prototype = ret0
	}

	return r.Prototype, nil

}

// GetCode returns reference to the code
func (r *RootDomain) GetCode() (insolar.Reference, error) {
	if r.Code.IsEmpty() {
		ret := [2]interface{}{}
		var ret0 insolar.Reference
		ret[0] = &ret0
		var ret1 *foundation.Error
		ret[1] = &ret1

		res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "GetCode", make([]byte, 0), *PrototypeReference)
		if err != nil {
			return ret0, err
		}

		err = common.CurrentProxyCtx.Deserialize(res, &ret)
		if err != nil {
			return ret0, err
		}

		if ret1 != nil {
			return ret0, ret1
		}

		r.Code = ret0
	}

	return r.Code, nil
}

// GetMDAdminMemberRef is proxy generated method
func (r *RootDomain) GetMDAdminMemberRef() (*insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 *insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "GetMDAdminMemberRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetMDAdminMemberRefNoWait is proxy generated method
func (r *RootDomain) GetMDAdminMemberRefNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "GetMDAdminMemberRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetMDAdminMemberRefAsImmutable is proxy generated method
func (r *RootDomain) GetMDAdminMemberRefAsImmutable() (*insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 *insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "GetMDAdminMemberRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetMDWalletRef is proxy generated method
func (r *RootDomain) GetMDWalletRef() (*insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 *insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "GetMDWalletRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetMDWalletRefNoWait is proxy generated method
func (r *RootDomain) GetMDWalletRefNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "GetMDWalletRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetMDWalletRefAsImmutable is proxy generated method
func (r *RootDomain) GetMDWalletRefAsImmutable() (*insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 *insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "GetMDWalletRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetOracleMembers is proxy generated method
func (r *RootDomain) GetOracleMembers() (map[string]insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 map[string]insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "GetOracleMembers", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetOracleMembersNoWait is proxy generated method
func (r *RootDomain) GetOracleMembersNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "GetOracleMembers", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetOracleMembersAsImmutable is proxy generated method
func (r *RootDomain) GetOracleMembersAsImmutable() (map[string]insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 map[string]insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "GetOracleMembers", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetRootMemberRef is proxy generated method
func (r *RootDomain) GetRootMemberRef() (*insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 *insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "GetRootMemberRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetRootMemberRefNoWait is proxy generated method
func (r *RootDomain) GetRootMemberRefNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "GetRootMemberRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetRootMemberRefAsImmutable is proxy generated method
func (r *RootDomain) GetRootMemberRefAsImmutable() (*insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 *insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "GetRootMemberRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetNodeDomainRef is proxy generated method
func (r *RootDomain) GetNodeDomainRef() (insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "GetNodeDomainRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetNodeDomainRefNoWait is proxy generated method
func (r *RootDomain) GetNodeDomainRefNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "GetNodeDomainRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetNodeDomainRefAsImmutable is proxy generated method
func (r *RootDomain) GetNodeDomainRefAsImmutable() (insolar.Reference, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "GetNodeDomainRef", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// Info is proxy generated method
func (r *RootDomain) Info() (interface{}, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 interface{}
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "Info", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// InfoNoWait is proxy generated method
func (r *RootDomain) InfoNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "Info", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// InfoAsImmutable is proxy generated method
func (r *RootDomain) InfoAsImmutable() (interface{}, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 interface{}
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "Info", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// AddBurnAddresses is proxy generated method
func (r *RootDomain) AddBurnAddresses(burnAddresses []string) error {
	var args [1]interface{}
	args[0] = burnAddresses

	var argsSerialized []byte

	ret := [1]interface{}{}
	var ret0 *foundation.Error
	ret[0] = &ret0

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "AddBurnAddresses", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return err
	}

	if ret0 != nil {
		return ret0
	}
	return nil
}

// AddBurnAddressesNoWait is proxy generated method
func (r *RootDomain) AddBurnAddressesNoWait(burnAddresses []string) error {
	var args [1]interface{}
	args[0] = burnAddresses

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "AddBurnAddresses", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// AddBurnAddressesAsImmutable is proxy generated method
func (r *RootDomain) AddBurnAddressesAsImmutable(burnAddresses []string) error {
	var args [1]interface{}
	args[0] = burnAddresses

	var argsSerialized []byte

	ret := [1]interface{}{}
	var ret0 *foundation.Error
	ret[0] = &ret0

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "AddBurnAddresses", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return err
	}

	if ret0 != nil {
		return ret0
	}
	return nil
}

// AddBurnAddress is proxy generated method
func (r *RootDomain) AddBurnAddress(burnAddress string) error {
	var args [1]interface{}
	args[0] = burnAddress

	var argsSerialized []byte

	ret := [1]interface{}{}
	var ret0 *foundation.Error
	ret[0] = &ret0

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "AddBurnAddress", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return err
	}

	if ret0 != nil {
		return ret0
	}
	return nil
}

// AddBurnAddressNoWait is proxy generated method
func (r *RootDomain) AddBurnAddressNoWait(burnAddress string) error {
	var args [1]interface{}
	args[0] = burnAddress

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "AddBurnAddress", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// AddBurnAddressAsImmutable is proxy generated method
func (r *RootDomain) AddBurnAddressAsImmutable(burnAddress string) error {
	var args [1]interface{}
	args[0] = burnAddress

	var argsSerialized []byte

	ret := [1]interface{}{}
	var ret0 *foundation.Error
	ret[0] = &ret0

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "AddBurnAddress", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return err
	}

	if ret0 != nil {
		return ret0
	}
	return nil
}

// GetBurnAddress is proxy generated method
func (r *RootDomain) GetBurnAddress() (string, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 string
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "GetBurnAddress", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetBurnAddressNoWait is proxy generated method
func (r *RootDomain) GetBurnAddressNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "GetBurnAddress", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetBurnAddressAsImmutable is proxy generated method
func (r *RootDomain) GetBurnAddressAsImmutable() (string, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 string
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "GetBurnAddress", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// AddNewMemberToMaps is proxy generated method
func (r *RootDomain) AddNewMemberToMaps(publicKey string, burnAddress string, memberRef insolar.Reference) error {
	var args [3]interface{}
	args[0] = publicKey
	args[1] = burnAddress
	args[2] = memberRef

	var argsSerialized []byte

	ret := [1]interface{}{}
	var ret0 *foundation.Error
	ret[0] = &ret0

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "AddNewMemberToMaps", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return err
	}

	if ret0 != nil {
		return ret0
	}
	return nil
}

// AddNewMemberToMapsNoWait is proxy generated method
func (r *RootDomain) AddNewMemberToMapsNoWait(publicKey string, burnAddress string, memberRef insolar.Reference) error {
	var args [3]interface{}
	args[0] = publicKey
	args[1] = burnAddress
	args[2] = memberRef

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "AddNewMemberToMaps", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// AddNewMemberToMapsAsImmutable is proxy generated method
func (r *RootDomain) AddNewMemberToMapsAsImmutable(publicKey string, burnAddress string, memberRef insolar.Reference) error {
	var args [3]interface{}
	args[0] = publicKey
	args[1] = burnAddress
	args[2] = memberRef

	var argsSerialized []byte

	ret := [1]interface{}{}
	var ret0 *foundation.Error
	ret[0] = &ret0

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "AddNewMemberToMaps", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return err
	}

	if ret0 != nil {
		return ret0
	}
	return nil
}

// GetMemberByBurnAddress is proxy generated method
func (r *RootDomain) GetMemberByBurnAddress(burnAddress string) (insolar.Reference, error) {
	var args [1]interface{}
	args[0] = burnAddress

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, "GetMemberByBurnAddress", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetMemberByBurnAddressNoWait is proxy generated method
func (r *RootDomain) GetMemberByBurnAddressNoWait(burnAddress string) error {
	var args [1]interface{}
	args[0] = burnAddress

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, "GetMemberByBurnAddress", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetMemberByBurnAddressAsImmutable is proxy generated method
func (r *RootDomain) GetMemberByBurnAddressAsImmutable(burnAddress string) (insolar.Reference, error) {
	var args [1]interface{}
	args[0] = burnAddress

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 insolar.Reference
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, "GetMemberByBurnAddress", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = common.CurrentProxyCtx.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}
