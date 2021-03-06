package claims

import (
	"encoding/hex"
	"testing"

	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/iden3/go-iden3-core/testgen"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
)

func testClaimAuthorizeKSignBabyJub(t *testing.T, i, testKey string) {
	// Create new claim
	var k babyjub.PrivateKey
	hexK := testgen.GetTestValue(i + testKey).(string)
	if _, err := hex.Decode(k[:], []byte(hexK)); err != nil {
		panic(err)
	}
	pk := k.Public()
	c0 := NewClaimAuthorizeKSignBabyJub(pk, 5678)
	assert.True(t, merkletree.CheckEntryInField(*c0.Entry()))
	c0.Version = 1
	e := c0.Entry()
	// Check claim against test vector
	testgen.CheckTestValue(t, "ClaimAuthorizeKSignBabyJub"+i+"_HIndex", e.HIndex().Hex())
	testgen.CheckTestValue(t, "ClaimAuthorizeKSignBabyJub"+i+"_HValue", e.HValue().Hex())
	testgen.CheckTestValue(t, "ClaimAuthorizeKSignBabyJub"+i+"_dataString", e.Data.String())
	dataTestOutput(&e.Data)
	c1 := NewClaimAuthorizeKSignBabyJubFromEntry(e)
	c2, err := NewClaimFromEntry(e)
	assert.Nil(t, err)
	assert.Equal(t, c0, c1)
	assert.Equal(t, c0, c2)
	assert.True(t, merkletree.CheckEntryInField(*e))

	// revocation nonce
	c3 := NewClaimAuthorizeKSignBabyJub(pk, 3)
	assert.Equal(t, c3.RevocationNonce, uint32(3))
	c3.Version = 1
	c1.RevocationNonce = 3
	assert.Equal(t, c3, c1)
}

func TestClaimAuthorizeKSignBabyJub(t *testing.T) {
	testClaimAuthorizeKSignBabyJub(t, "0", "_privateKey")
	testClaimAuthorizeKSignBabyJub(t, "1", "_privateKey")
}

func TestRandomClaimAuthorizeKSignBabyJub(t *testing.T) {
	for i := 0; i < 100; i++ {
		k := babyjub.NewRandPrivKey()
		pk := k.Public()

		c0 := NewClaimAuthorizeKSignBabyJub(pk, 0)
		assert.True(t, merkletree.CheckEntryInField(*c0.Entry()))
		c0.Version = 9999
		e := c0.Entry()
		c1 := NewClaimAuthorizeKSignBabyJubFromEntry(e)
		c2, err := NewClaimFromEntry(e)
		assert.Nil(t, err)
		assert.Equal(t, c0, c1)
		assert.Equal(t, c0, c2)
		assert.True(t, merkletree.CheckEntryInField(*e))
	}
}
