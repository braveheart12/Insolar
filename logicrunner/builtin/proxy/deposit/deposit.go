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

package deposit

import (
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/logicrunner/common"
	"github.com/insolar/insolar/logicrunner/goplugin/foundation"
	"time"
)

type status string

// PrototypeReference to prototype of this contract
// error checking hides in generator
var PrototypeReference, _ = insolar.NewReferenceFromBase58("111A7ctasuNUug8BoK4VJNuAFJ73rnH8bH5zqd5HrDj.11111111111111111111111111111111")

// Deposit holds proxy type
type Deposit struct {
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
func (r *ContractConstructorHolder) AsChild(objRef insolar.Reference) (*Deposit, error) {
	ref, err := common.CurrentProxyCtx.SaveAsChild(objRef, *PrototypeReference, r.constructorName, r.argsSerialized)
	if err != nil {
		return nil, err
	}
	return &Deposit{Reference: ref}, nil
}

// AsDelegate saves object as delegate
func (r *ContractConstructorHolder) AsDelegate(objRef insolar.Reference) (*Deposit, error) {
	ref, err := common.CurrentProxyCtx.SaveAsDelegate(objRef, *PrototypeReference, r.constructorName, r.argsSerialized)
	if err != nil {
		return nil, err
	}
	return &Deposit{Reference: ref}, nil
}

// GetObject returns proxy object
func GetObject(ref insolar.Reference) (r *Deposit) {
	return &Deposit{Reference: ref}
}

// GetPrototype returns reference to the prototype
func GetPrototype() insolar.Reference {
	return *PrototypeReference
}

// GetImplementationFrom returns proxy to delegate of given type
func GetImplementationFrom(object insolar.Reference) (*Deposit, error) {
	ref, err := common.CurrentProxyCtx.GetDelegate(object, *PrototypeReference)
	if err != nil {
		return nil, err
	}
	return GetObject(ref), nil
}

// New is constructor
func New(migrationDaemonConfirms map[insolar.Reference]bool, txHash string, amount string, holdReleaseDate time.Time) *ContractConstructorHolder {
	var args [4]interface{}
	args[0] = migrationDaemonConfirms
	args[1] = txHash
	args[2] = amount
	args[3] = holdReleaseDate

	var argsSerialized []byte
	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		panic(err)
	}

	return &ContractConstructorHolder{constructorName: "New", argsSerialized: argsSerialized}
}

// GetReference returns reference of the object
func (r *Deposit) GetReference() insolar.Reference {
	return r.Reference
}

// GetPrototype returns reference to the code
func (r *Deposit) GetPrototype() (insolar.Reference, error) {
	if r.Prototype.IsEmpty() {
		ret := [2]interface{}{}
		var ret0 insolar.Reference
		ret[0] = &ret0
		var ret1 *foundation.Error
		ret[1] = &ret1

		res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, false, "GetPrototype", make([]byte, 0), *PrototypeReference)
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
func (r *Deposit) GetCode() (insolar.Reference, error) {
	if r.Code.IsEmpty() {
		ret := [2]interface{}{}
		var ret0 insolar.Reference
		ret[0] = &ret0
		var ret1 *foundation.Error
		ret[1] = &ret1

		res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, false, "GetCode", make([]byte, 0), *PrototypeReference)
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

// GetTxHash is proxy generated method
func (r *Deposit) GetTxHash() (string, error) {
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

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, false, "GetTxHash", argsSerialized, *PrototypeReference)
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

// GetTxHashNoWait is proxy generated method
func (r *Deposit) GetTxHashNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, false, "GetTxHash", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetTxHashAsImmutable is proxy generated method
func (r *Deposit) GetTxHashAsImmutable() (string, error) {
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

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, false, "GetTxHash", argsSerialized, *PrototypeReference)
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

// GetAmount is proxy generated method
func (r *Deposit) GetAmount() (string, error) {
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

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, false, "GetAmount", argsSerialized, *PrototypeReference)
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

// GetAmountNoWait is proxy generated method
func (r *Deposit) GetAmountNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, false, "GetAmount", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetAmountAsImmutable is proxy generated method
func (r *Deposit) GetAmountAsImmutable() (string, error) {
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

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, false, "GetAmount", argsSerialized, *PrototypeReference)
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

// MapMarshal is proxy generated method
func (r *Deposit) MapMarshal() (map[string]string, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 map[string]string
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, false, "MapMarshal", argsSerialized, *PrototypeReference)
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

// MapMarshalNoWait is proxy generated method
func (r *Deposit) MapMarshalNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, false, "MapMarshal", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// MapMarshalAsImmutable is proxy generated method
func (r *Deposit) MapMarshalAsImmutable() (map[string]string, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 map[string]string
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, false, "MapMarshal", argsSerialized, *PrototypeReference)
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

// Confirm is proxy generated method
func (r *Deposit) Confirm(migrationDaemon insolar.Reference, txHash string, amountStr string) (uint, error) {
	var args [3]interface{}
	args[0] = migrationDaemon
	args[1] = txHash
	args[2] = amountStr

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 uint
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, false, false, "Confirm", argsSerialized, *PrototypeReference)
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

// ConfirmNoWait is proxy generated method
func (r *Deposit) ConfirmNoWait(migrationDaemon insolar.Reference, txHash string, amountStr string) error {
	var args [3]interface{}
	args[0] = migrationDaemon
	args[1] = txHash
	args[2] = amountStr

	var argsSerialized []byte

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = common.CurrentProxyCtx.RouteCall(r.Reference, false, false, false, "Confirm", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// ConfirmAsImmutable is proxy generated method
func (r *Deposit) ConfirmAsImmutable(migrationDaemon insolar.Reference, txHash string, amountStr string) (uint, error) {
	var args [3]interface{}
	args[0] = migrationDaemon
	args[1] = txHash
	args[2] = amountStr

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 uint
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := common.CurrentProxyCtx.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := common.CurrentProxyCtx.RouteCall(r.Reference, true, true, false, "Confirm", argsSerialized, *PrototypeReference)
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
