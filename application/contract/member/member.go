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

package member

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/insolar/insolar/application/contract/member/signer"
	"github.com/insolar/insolar/application/proxy/deposit"
	"github.com/insolar/insolar/application/proxy/member"
	"github.com/insolar/insolar/application/proxy/nodedomain"
	"github.com/insolar/insolar/application/proxy/rootdomain"
	"github.com/insolar/insolar/application/proxy/wallet"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/logicrunner/goplugin/foundation"
)

type Member struct {
	foundation.BaseContract
	Name      string
	EthAddr   string
	PublicKey string
}

func (m *Member) GetName() (string, error) {
	return m.Name, nil
}

func (m *Member) GetEthAddr() (string, error) {
	return m.EthAddr, nil
}

func (m *Member) SetEthAddr(ethAddr string) error {
	m.EthAddr = ethAddr
	return nil
}

var INSATTR_GetPublicKey_API = true

func (m *Member) GetPublicKey() (string, error) {
	return m.PublicKey, nil
}

func New(ethAddr string, key string) (*Member, error) {
	return &Member{
		EthAddr:   ethAddr,
		PublicKey: key,
	}, nil
}

func NewOracleMember(name string, key string) (*Member, error) {
	return &Member{
		Name:      name,
		PublicKey: key,
	}, nil
}

func (m *Member) verifySig(method string, params []byte, seed []byte, sign []byte) error {
	args, err := insolar.MarshalArgs(m.GetReference(), method, params, seed)
	if err != nil {
		return fmt.Errorf("[ verifySig ] Can't MarshalArgs: %s", err.Error())
	}

	key, err := m.GetPublicKey()
	if err != nil {
		return fmt.Errorf("[ verifySig ]: %s", err.Error())
	}

	publicKey, err := foundation.ImportPublicKey(key)
	if err != nil {
		return fmt.Errorf("[ verifySig ] Invalid public key")
	}

	verified := foundation.Verify(args, sign, publicKey)
	if !verified {
		return fmt.Errorf("[ verifySig ] Incorrect signature")
	}
	return nil
}

var INSATTR_Call_API = true

// Call method for authorized calls
func (m *Member) Call(rootDomainRef insolar.Reference, method string, params []byte, seed []byte, sign []byte) (interface{}, error) {

	switch method {
	case "CreateMember":
		return m.createMemberCall(rootDomainRef, params)
	}

	if err := m.verifySig(method, params, seed, sign); err != nil {
		return nil, fmt.Errorf("[ Call ]: %s", err.Error())
	}

	switch method {
	case "GetBalance":
		return m.getBalanceCall()
	case "Transfer":
		return m.transferCall(params)
	case "DumpUserInfo":
		return m.dumpUserInfoCall(rootDomainRef, params)
	case "DumpAllUsers":
		return m.dumpAllUsersCall(rootDomainRef)
	case "RegisterNode":
		return m.registerNodeCall(rootDomainRef, params)
	case "GetNodeRef":
		return m.getNodeRefCall(rootDomainRef, params)
	case "Migration":
		return m.migrationCall(rootDomainRef, params)
	case "AddBurnAddress":
		return m.AddBurnAddressCall(rootDomainRef, params)
	}
	return nil, &foundation.Error{S: "Unknown method"}
}

func (m *Member) createMemberCall(rdRef insolar.Reference, params []byte) (interface{}, error) {
	var key string
	if err := signer.UnmarshalParams(params, &key); err != nil {
		return nil, fmt.Errorf("[ createMemberCall ]: %s", err.Error())
	}

	return m.createMemberByKey(rdRef, key)
}

func (m *Member) createMemberByKey(rdRef insolar.Reference, key string) (interface{}, error) {

	rootDomain := rootdomain.GetObject(rdRef)
	ba, err := rootDomain.GetBurnAddress()
	if err != nil {
		return nil, fmt.Errorf("[ createMemberByKey ] Can't get burn address: %s", err.Error())
	}

	new, err := m.createMember(rdRef, ba, key)
	if err != nil {
		if e := rootDomain.AddBurnAddress(ba); e != nil {
			return nil, fmt.Errorf("[ createMemberByKey ] Can't add burn address back: %s; after error: %s", e.Error(), err.Error())
		}
		return nil, fmt.Errorf("[ createMemberByKey ] Can't create member: %s", err.Error())
	}

	if err = rootDomain.AddNewMemberToMaps(key, ba, new.Reference); err != nil {
		return nil, fmt.Errorf("[ createMemberByKey ] Can't add new member to maps: %s", err.Error())
	}

	return new.Reference.String(), nil
}

func (m *Member) createMember(rdRef insolar.Reference, ethAddr string, key string) (*member.Member, error) {
	if key == "" {
		return nil, fmt.Errorf("[ createMember ] Key is not valid")
	}

	memberHolder := member.New(ethAddr, key)
	new, err := memberHolder.AsChild(rdRef)
	if err != nil {
		return nil, fmt.Errorf("[ createMember ] Can't save as child: %s", err.Error())
	}

	wHolder := wallet.New(big.NewInt(100).String())
	_, err = wHolder.AsDelegate(new.Reference)
	if err != nil {
		return nil, fmt.Errorf("[ createMember ] Can't save as delegate: %s", err.Error())
	}

	return new, nil
}

func (m *Member) getDeposits() ([]map[string]string, error) {

	iterator, err := m.NewChildrenTypedIterator(deposit.GetPrototype())
	if err != nil {
		return nil, fmt.Errorf("[ getDeposits ] Can't get children: %s", err.Error())
	}

	result := []map[string]string{}
	for iterator.HasNext() {
		cref, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("[ getDeposits ] Can't get next child: %s", err.Error())
		}

		if !cref.IsEmpty() {
			d := deposit.GetObject(cref)

			m, err := d.MapMarshal()
			if err != nil {
				return nil, fmt.Errorf("[ getDeposits ] Can't marshal deposit to map: %s", err.Error())
			}

			result = append(result, m)
		}
	}

	return result, nil
}

type BalanceWithDeposits struct {
	Balance string
	//Deposits []map[string]string
}

func (m *Member) getBalanceCall() (interface{}, error) {
	w, err := wallet.GetImplementationFrom(m.GetReference())
	if err != nil {
		return nil, fmt.Errorf("[ getBalanceCall ] Can't get implementation: %s", err.Error())
	}
	b, err := w.GetBalance()
	if err != nil {
		return nil, fmt.Errorf("[ getBalanceCall ] Can't get balance: %s", err.Error())
	}
	//d, err := m.getDeposits()
	//if err != nil {
	//	return nil, fmt.Errorf("[ getBalanceCall ] Can't get deposits: %s", err.Error())
	//}

	balanceWithDepositsMarshaled, err := json.Marshal(BalanceWithDeposits{
		Balance: b,
		//Deposits: d,
	})
	if err != nil {
		return nil, fmt.Errorf("[ getBalanceCall ] Can't marshal: %s", err.Error())
	}

	return balanceWithDepositsMarshaled, nil
}

func parseTimeStamp(timeStr string) (time.Time, error) {

	i, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time.Unix(0, 0), errors.New("Can't parse time ")
	}
	return time.Unix(i, 0), nil
}

func (m *Member) transferCall(params []byte) (interface{}, error) {
	var toMemberStr string
	var amount string
	if err := signer.UnmarshalParams(params, &amount, &toMemberStr); err != nil {
		return nil, fmt.Errorf("[ transferCall ] Can't unmarshal params: %s", err.Error())
	}

	toMember, err := insolar.NewReferenceFromBase58(toMemberStr)
	if err != nil {
		return nil, fmt.Errorf("[ transferCall ] Failed to parse 'to' param: %s", err.Error())
	}
	if m.GetReference() == *toMember {
		return nil, fmt.Errorf("[ transferCall ] Recipient must be different from the sender")
	}

	return m.transfer(amount, toMember)
}

func (m *Member) transfer(amount string, toMember *insolar.Reference) (interface{}, error) {

	w, err := wallet.GetImplementationFrom(m.GetReference())
	if err != nil {
		return nil, fmt.Errorf("[ transfer ] Can't get wallet implementation of sender: %s", err.Error())
	}

	return nil, w.Transfer(amount, toMember)
}

func (m *Member) registerNodeCall(rdRef insolar.Reference, params []byte) (interface{}, error) {
	var publicKey string
	var role string
	if err := signer.UnmarshalParams(params, &publicKey, &role); err != nil {
		return nil, fmt.Errorf("[ registerNodeCall ] Can't unmarshal params: %s", err.Error())
	}

	rootDomain := rootdomain.GetObject(rdRef)
	nodeDomainRef, err := rootDomain.GetNodeDomainRef()
	if err != nil {
		return nil, fmt.Errorf("[ registerNodeCall ] %s", err.Error())
	}

	nd := nodedomain.GetObject(nodeDomainRef)
	cert, err := nd.RegisterNode(publicKey, role)
	if err != nil {
		return nil, fmt.Errorf("[ registerNodeCall ] Problems with RegisterNode: %s", err.Error())
	}

	return string(cert), nil
}

func (m *Member) getNodeRefCall(rdRef insolar.Reference, params []byte) (interface{}, error) {
	var publicKey string
	if err := signer.UnmarshalParams(params, &publicKey); err != nil {
		return nil, fmt.Errorf("[ getNodeRefCall ] Can't unmarshal params: %s", err.Error())
	}

	rootDomain := rootdomain.GetObject(rdRef)
	nodeDomainRef, err := rootDomain.GetNodeDomainRef()
	if err != nil {
		return nil, fmt.Errorf("[ getNodeRefCall ] Can't get nodeDmainRef: %s", err.Error())
	}

	nd := nodedomain.GetObject(nodeDomainRef)
	nodeRef, err := nd.GetNodeRefByPK(publicKey)
	if err != nil {
		return nil, fmt.Errorf("[ getNodeRefCall ] NetworkNode not found: %s", err.Error())
	}

	return nodeRef, nil
}

func (mdMember *Member) migration(rdRef insolar.Reference, txHash string, burnAddress string, amount big.Int, unHoldDate time.Time) (string, error) {
	rd := rootdomain.GetObject(rdRef)

	// Get oracle members
	oracleMembers, err := rd.GetOracleMembers()
	if err != nil {
		return "", fmt.Errorf("[ migration ] Can't get oracles map: %s", err.Error())
	}
	// Check that caller is oracle
	if _, ok := oracleMembers[mdMember.Name]; !ok {
		return "", fmt.Errorf("[ migration ] This oracle is not in the list")
	}

	// Get member by burn address
	mRef, err := rd.GetMemberByBurnAddress(burnAddress)
	if err != nil {
		return "", fmt.Errorf("[ migration ] Failed to get member by burn address")
	}
	m := member.GetObject(mRef)

	// Find deposit for txHash
	found, txDeposit, err := m.FindDeposit(txHash, amount)
	if err != nil {
		return "", fmt.Errorf("[ migration ] Can't get deposit: %s", err.Error())
	}

	// If deposit doesn't exist - create new deposit
	if !found {
		oracleConfirms := map[string]bool{}
		for name, _ := range oracleMembers {
			oracleConfirms[name] = false
		}
		dHolder := deposit.New(oracleConfirms, txHash, amount, unHoldDate)
		txDepositP, err := dHolder.AsDelegate(mRef)
		if err != nil {
			return "", fmt.Errorf("[ migration ] Can't save as delegate: %s", err.Error())
		}
		txDeposit = *txDepositP
	}

	// Confirm tx by oracle
	confirms, err := txDeposit.Confirm(mdMember.Name, txHash, amount)
	if err != nil {
		return "", fmt.Errorf("[ migration ] Confirmed failed: %s", err.Error())
	}

	//if allConfirmed {
	//	w, err := wallet.GetImplementationFrom(insAddr)
	//	if err != nil {
	//		wHolder := wallet.New(0)
	//		w, err = wHolder.AsDelegate(insAddr)
	//		if err != nil {
	//			return "", fmt.Errorf("[ migration ] Can't save as delegate: %s", err.Error())
	//		}
	//	}
	//
	//	getMdWallet := func() (*wallet.Wallet, error) {
	//		mdWalletRef, err := rd.GetMDWalletRef()
	//		if err != nil {
	//			return nil, fmt.Errorf("[ migration ] Can't get md wallet ref: %s", err.Error())
	//		}
	//		mdWallet := wallet.GetObject(*mdWalletRef)
	//
	//		return mdWallet, nil
	//	}
	//	mdWallet, err := getMdWallet()
	//	if err != nil {
	//		return "", fmt.Errorf("[ migration ] Can't get mdWallet: %s", err.Error())
	//	}
	//
	//	err = mdWallet.Transfer(amount, &w.Reference)
	//	if err != nil {
	//		return "", fmt.Errorf("[ migration ] Can't transfer: %s", err.Error())
	//	}
	//
	//}
	//
	//return insAddr.String(), nil
	return strconv.Itoa(int(confirms)), nil
}

func (mdMember *Member) migrationCall(rdRef insolar.Reference, params []byte) (string, error) {
	if mdMember.Name == "" {
		return "", fmt.Errorf("[ migrationCall ] Only oracles can call migrationCall")
	}

	var txHash, burnAddress, currentDate, inAmount string
	if err := signer.UnmarshalParams(params, &txHash, &burnAddress, &inAmount, &currentDate); err != nil {
		return "", fmt.Errorf("[ migrationCall ] Can't unmarshal params: %s", err.Error())
	}

	amount := new(big.Int)
	amount, ok := amount.SetString(inAmount, 10)
	if !ok {
		return "", fmt.Errorf("[ migrationCall ] Failed to parse amount")
	}

	unHoldDate, err := parseTimeStamp(currentDate)
	if err != nil {
		return "", fmt.Errorf("[ migrationCall ] Failed to parse unHoldDate: %s", err.Error())
	}

	return mdMember.migration(rdRef, txHash, burnAddress, *amount, unHoldDate)
}

func (m *Member) FindDeposit(txHash string, amount big.Int) (bool, deposit.Deposit, error) {
	iterator, err := m.NewChildrenTypedIterator(deposit.GetPrototype())
	if err != nil {
		return false, deposit.Deposit{}, fmt.Errorf("[ findDeposit ] Can't get children: %s", err.Error())
	}

	for iterator.HasNext() {
		cref, err := iterator.Next()
		if err != nil {
			return false, deposit.Deposit{}, fmt.Errorf("[ findDeposit ] Can't get next child: %s", err.Error())
		}

		if !cref.IsEmpty() {
			d := deposit.GetObject(cref)
			th, err := d.GetTxHash()
			if err != nil {
				return false, deposit.Deposit{}, fmt.Errorf("[ findDeposit ] Can't get tx hash: %s", err.Error())
			}
			a, err := d.GetAmount()
			if err != nil {
				return false, deposit.Deposit{}, fmt.Errorf("[ findDeposit ] Can't get amount: %s", err.Error())
			}

			if txHash == th {
				if (&amount).Cmp(&a) == 0 {
					return true, *d, nil
				}
			}
		}
	}

	return false, deposit.Deposit{}, nil
}

func (m *Member) dumpUserInfoCall(rdRef insolar.Reference, params []byte) (interface{}, error) {
	var userRefIn string
	if err := signer.UnmarshalParams(params, &userRefIn); err != nil {
		return nil, fmt.Errorf("[ dumpUserInfoCall ] Can't unmarshal params: %s", err.Error())
	}
	userRef, err := insolar.NewReferenceFromBase58(userRefIn)
	if err != nil {
		return nil, fmt.Errorf("[ dumpUserInfoCall ] Failed to parse 'inInsAddr' param: %s", err.Error())
	}

	rootDomain := rootdomain.GetObject(rdRef)
	rootMember, err := rootDomain.GetRootMemberRef()
	if err != nil {
		return nil, fmt.Errorf("[ dumpUserInfoCall ] Can't get root member: %s", err.Error())
	}
	if *userRef != m.GetReference() && m.GetReference() != *rootMember {
		return nil, fmt.Errorf("[ dumpUserInfoCall ] You can dump only yourself")
	}

	return m.DumpUserInfo(rdRef, *userRef)
}

func (m *Member) dumpAllUsersCall(rdRef insolar.Reference) (interface{}, error) {
	rootDomain := rootdomain.GetObject(rdRef)
	rootMember, err := rootDomain.GetRootMemberRef()
	if err != nil {
		return nil, fmt.Errorf("[ dumpAllUsersCall ] Can't get root member: %s", err.Error())
	}
	if m.GetReference() != *rootMember {
		return nil, fmt.Errorf("[ dumpAllUsersCall ] You can dump only yourself")
	}

	return m.DumpAllUsers(rdRef)
}

func (rootMember *Member) getUserInfoMap(m *member.Member) (map[string]interface{}, error) {
	w, err := wallet.GetImplementationFrom(m.GetReference())
	if err != nil {
		return nil, fmt.Errorf("[ getUserInfoMap ] Can't get implementation: %s", err.Error())
	}

	name, err := m.GetName()
	if err != nil {
		return nil, fmt.Errorf("[ getUserInfoMap ] Can't get name: %s", err.Error())
	}

	ethAddr, err := m.GetEthAddr()
	if err != nil {
		return nil, fmt.Errorf("[ getUserInfoMap ] Can't get name: %s", err.Error())
	}

	balance, err := w.GetBalance()
	if err != nil {
		return nil, fmt.Errorf("[ getUserInfoMap ] Can't get total balance: %s", err.Error())
	}
	return map[string]interface{}{
		"name":    name,
		"ethAddr": ethAddr,
		"balance": balance,
	}, nil
}

// DumpUserInfo processes dump user info request
func (m *Member) DumpUserInfo(rdRef insolar.Reference, userRef insolar.Reference) ([]byte, error) {

	user := member.GetObject(userRef)
	res, err := m.getUserInfoMap(user)
	if err != nil {
		return nil, fmt.Errorf("[ DumpUserInfo ] Problem with making request: %s", err.Error())
	}

	return json.Marshal(res)
}

// DumpAllUsers processes dump all users request
func (rootMember *Member) DumpAllUsers(rdRef insolar.Reference) ([]byte, error) {

	res := []map[string]interface{}{}

	rootDomain := rootdomain.GetObject(rdRef)
	iterator, err := rootDomain.DumpAllUsers()
	if err != nil {
		return nil, fmt.Errorf("[ DumpAllUsers ] Can't get children: %s", err.Error())
	}

	for iterator.HasNext() {
		cref, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("[ DumpAllUsers ] Can't get next child: %s", err.Error())
		}

		if !cref.IsEmpty() {
			m := member.GetObject(cref)
			userInfo, err := rootMember.getUserInfoMap(m)
			if err != nil {
				return nil, fmt.Errorf("[ DumpAllUsers ] Problem with making request: %s", err.Error())
			}
			res = append(res, userInfo)
		}
	}
	resJSON, _ := json.Marshal(res)
	return resJSON, nil
}

func (mdAdminMember *Member) AddBurnAddressCall(rdRef insolar.Reference, params []byte) (interface{}, error) {

	rootDomain := rootdomain.GetObject(rdRef)
	mdAdminRef, err := rootDomain.GetMDAdminMemberRef()
	if err != nil {
		return nil, fmt.Errorf("[ AddBurnAddressCall ] Can't get migration deamon admin reference from root domain: %s", err.Error())
	}

	if mdAdminMember.GetReference() != *mdAdminRef {
		return nil, fmt.Errorf("[ AddBurnAddressCall ] Only migration deamon admin can call this method")
	}

	var burnAddress string
	if err := signer.UnmarshalParams(params, &burnAddress); err != nil {
		return nil, fmt.Errorf("[ AddBurnAddressCall ] Can't unmarshal params: %s", err.Error())
	}

	err = rootDomain.AddBurnAddress(burnAddress)
	if err != nil {
		return nil, fmt.Errorf("[ AddBurnAddressCall ] Can't add burn address: %s", err.Error())
	}

	return nil, nil
}
