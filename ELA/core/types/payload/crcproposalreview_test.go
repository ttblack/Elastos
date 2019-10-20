// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

package payload

import (
	"bytes"
	"crypto/rand"
	mathRand "math/rand"
	"testing"

	"github.com/elastos/Elastos.ELA/common"
	"github.com/stretchr/testify/assert"
)

func TestCRCProposalReview_Deserialize(t *testing.T) {
	proposalReview1 := randomCRCProposalReviewPayload()

	buf := new(bytes.Buffer)
	proposalReview1.Serialize(buf, CRCProposalReviewVersion)

	proposalReview2 := &CRCProposalReview{}
	proposalReview2.Deserialize(buf, CRCProposalReviewVersion)

	assert.True(t, crcProposalReviewPayloadEqual(proposalReview1, proposalReview2))
}

func crcProposalReviewPayloadEqual(payload1 *CRCProposalReview,
	payload2 *CRCProposalReview) bool {
	if !payload1.ProposalHash.IsEqual(payload2.ProposalHash) ||
		payload1.VoteResult != payload2.VoteResult ||
		!payload1.DID.IsEqual(payload2.DID) ||
		!bytes.Equal(payload1.Sign, payload2.Sign) {
		return false
	}

	return true
}

func randomCRCProposalReviewPayload() *CRCProposalReview {
	return &CRCProposalReview{
		ProposalHash: *randomUint256(),
		VoteResult:   VoteResult(mathRand.Int() % (int(Abstain) + 1)),
		DID:          *randomUint168(),
		Sign:         randomBytes(65),
	}
}
func randomUint168() *common.Uint168 {
	randBytes := make([]byte, 21)
	rand.Read(randBytes)
	result, _ := common.Uint168FromBytes(randBytes)

	return result
}
func randomUint256() *common.Uint256 {
	randBytes := make([]byte, 32)
	rand.Read(randBytes)

	result, _ := common.Uint256FromBytes(randBytes)
	return result
}

func randomBytes(len int) []byte {
	a := make([]byte, len)
	rand.Read(a)
	return a
}
