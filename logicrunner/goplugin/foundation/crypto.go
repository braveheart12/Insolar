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

package foundation

import (
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"

	"github.com/insolar/x-crypto/ecdsa"
	"github.com/insolar/x-crypto/sha256"
	"github.com/insolar/x-crypto/x509"
)

// UnmarshalSig parses the two integer components of an ASN.1-encoded ECDSA signature.
func UnmarshalSig(b []byte) (r, s *big.Int, err error) {
	var ecsdaSig struct {
		R, S *big.Int
	}
	_, err = asn1.Unmarshal(b, &ecsdaSig)
	if err != nil {
		return nil, nil, err
	}
	return ecsdaSig.R, ecsdaSig.S, nil
}

// VerifySignature used for checking the signature using rawpublicpem and rawRequest.
// selfSigned flag need to compare public keys.
func VerifySignature(rawRequest []byte, signature string, key string, rawpublicpem string, selfSigned bool) error {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("cant decode signature %s", err.Error())
	}

	if key != rawpublicpem && !selfSigned {
		return fmt.Errorf("access denied. Key - %v", rawpublicpem)
	}

	blockPub, _ := pem.Decode([]byte(rawpublicpem))
	if blockPub == nil {
		return fmt.Errorf("problems with decoding. Key - %v", rawpublicpem)
	}
	x509EncodedPub := blockPub.Bytes
	publicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		return fmt.Errorf("problems with parsing. Key - %v", rawpublicpem)
	}

	hash := sha256.Sum256(rawRequest)
	r, s, err := UnmarshalSig(sig)
	if err != nil {
		return err
	}
	valid := ecdsa.Verify(publicKey.(*ecdsa.PublicKey), hash[:], r, s)
	if !valid {
		return fmt.Errorf("invalid signature")
	}

	return nil
}
