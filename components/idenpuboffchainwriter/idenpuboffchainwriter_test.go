package idenpuboffchainwriter

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/iden3/go-iden3-core/core/claims"
	"github.com/iden3/go-iden3-core/db"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/iden3/go-iden3-core/testgen"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// If generateTest is true, the checked values will be used to generate a test vector
var generateTest = false

func TestHttpPublicGetPublicData(t *testing.T) {
	// create RootsTree & RevocationsTree
	cltMt, err := merkletree.NewMerkleTree(db.NewMemoryStorage(), 140)
	assert.Nil(t, err)
	rotMt, err := merkletree.NewMerkleTree(db.NewMemoryStorage(), 140)
	assert.Nil(t, err)
	retMt, err := merkletree.NewMerkleTree(db.NewMemoryStorage(), 140)
	assert.Nil(t, err)

	// add some leafs to both MerkleTrees
	for i := 0; i < 10; i++ {
		rootBigInt, err := poseidon.HashBytes([]byte(strconv.Itoa(i)))
		assert.Nil(t, err)
		root := merkletree.BigIntToHash(rootBigInt)
		err = claims.AddLeafRootsTree(rotMt, &root)
		assert.Nil(t, err)

		nonce := uint32(i)
		version := uint32(i)
		err = claims.AddLeafRevocationsTree(retMt, nonce, version)
		assert.Nil(t, err)
	}

	testgen.CheckTestValue(t, "rootRootsTree1", rotMt.RootKey().Hex())
	testgen.CheckTestValue(t, "rootRevocationsTree1", retMt.RootKey().Hex())

	idenPubOffChainWriteHttp, err := NewIdenPubOffChainWriteHttp(&ConfigDefault, db.NewMemoryStorage(), rotMt, retMt)
	require.Nil(t, err)

	idenState := merkletree.HexStringToHash(testgen.GetTestValue("idenState0").(string))

	err = idenPubOffChainWriteHttp.Publish(&idenState, cltMt.RootKey(), retMt.RootKey(), rotMt.RootKey())
	assert.Nil(t, err)

	pubData, err := idenPubOffChainWriteHttp.GetPublicData(nil)
	assert.Nil(t, err)
	testgen.CheckTestValue(t, "rootRootsTree1", pubData.RootsTreeRoot.Hex())
	assert.Equal(t, rotMt.RootKey().Hex(), pubData.RootsTreeRoot.Hex())
	testgen.CheckTestValue(t, "rootRevocationsTree1", pubData.RevocationsTreeRoot.Hex())
	assert.Equal(t, retMt.RootKey().Hex(), pubData.RevocationsTreeRoot.Hex())
}

func initTest() {
	// Init test
	err := testgen.InitTest("idenpuboffchainwriter", generateTest)
	if err != nil {
		fmt.Println("error initializing test data:", err)
		return
	}
	// Add input data to the test vector
	if generateTest {
		idenState0, err := poseidon.HashBytes([]byte("idenState0"))
		if err != nil {
			panic(err)
		}
		testgen.SetTestValue("idenState0", merkletree.BigIntToHash(idenState0).Hex())
		root0, err := poseidon.HashBytes([]byte("root0"))
		if err != nil {
			panic(err)
		}
		testgen.SetTestValue("root0", merkletree.BigIntToHash(root0).Hex())
		testgen.SetTestValue("nonce0", float64(5))
		testgen.SetTestValue("version0", float64(5))
	}
}

func TestMain(m *testing.M) {
	initTest()
	result := m.Run()
	if err := testgen.StopTest(); err != nil {
		panic(fmt.Errorf("Error stopping test: %w", err))
	}
	os.Exit(result)
}
