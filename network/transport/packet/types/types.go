/*
 *    Copyright 2018 Insolar
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package types

//go:generate stringer -type=PacketType
type PacketType int

const (
	// TypePing is packet type for ping method.
	TypePing PacketType = iota + 1
	// TypeStore is packet type for store method.
	TypeStore
	// TypeFindHost is packet type for FindHost method.
	TypeFindHost
	// TypeFindValue is packet type for FindValue method.
	TypeFindValue
	// TypeRPC is packet type for RPC method.
	TypeRPC
	// TypeRelay is packet type for request target to be a relay.
	TypeRelay
	// TypeAuthentication is packet type for authentication between hosts.
	TypeAuthentication
	// TypeCheckOrigin is packet to check originality of some host.
	TypeCheckOrigin
	// TypeObtainIP is packet to get itself IP from another host.
	TypeObtainIP
	// TypeRelayOwnership is packet to say all other hosts that current host have a static IP.
	TypeRelayOwnership
	// TypeKnownOuterHosts is packet to say how much outer hosts current host know.
	TypeKnownOuterHosts
	// TypeCheckNodePriv is packet to check preset node privileges.
	TypeCheckNodePriv
	// TypeCascadeSend is the packet type for the cascade send message feature.
	TypeCascadeSend
	// TypePulse is packet type for the messages received from pulsars.
	TypePulse
	// TypeGetRandomHosts is packet type for the call to get random hosts of the DHT network.
	TypeGetRandomHosts
	// TypeGetNonce is packet to request a nonce to sign it.
	TypeGetNonce
	// TypeCheckSignedNonce is packet to check a signed nonce.
	TypeCheckSignedNonce
	// TypeExchangeUnsyncLists is packet type to exchange unsync lists during consensus
	TypeExchangeUnsyncLists
	// TypeExchangeUnsyncHash is packet type to exchange hash of merged unsync lists during consensus
	TypeExchangeUnsyncHash
	// TypeDisconnect is packet to disconnect from active list.
	TypeDisconnect
)