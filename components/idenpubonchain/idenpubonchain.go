package idenpubonchain

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/core/proof"
	"github.com/iden3/go-iden3-core/eth"
	"github.com/iden3/go-iden3-core/eth/contracts"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// IdenPubOnChainer is an interface that gives access to the IdenStates Smart Contract.
type IdenPubOnChainer interface {
	GetState(id *core.ID) (*proof.IdenStateData, error)
	GetStateByBlock(id *core.ID, blockN uint64) (*proof.IdenStateData, error)
	GetStateByTime(id *core.ID, blockTimestamp int64) (*proof.IdenStateData, error)
	SetState(id *core.ID, newState *merkletree.Hash, kOpProof []byte, stateTransitionProof []byte, signature *babyjub.SignatureComp) (*types.Transaction, error)
	InitState(id *core.ID, genesisState *merkletree.Hash, newState *merkletree.Hash, kOpProof []byte, stateTransitionProof []byte, signature *babyjub.SignatureComp) (*types.Transaction, error)
	// VerifyProofClaim(pc *proof.ProofClaim) (bool, error)
}

// ContractAddresses are the list of Smart Contract addresses used for the on chain identity state data.
type ContractAddresses struct {
	IdenStates common.Address
}

// IdenPubOnChain is the regular implementation of IdenPubOnChain
type IdenPubOnChain struct {
	client    *eth.Client2
	addresses ContractAddresses
}

// New creates a new IdenPubOnChain
func New(client *eth.Client2, addresses ContractAddresses) *IdenPubOnChain {
	return &IdenPubOnChain{
		client:    client,
		addresses: addresses,
	}
}

// GetState returns the Identity State Data of the given ID from the IdenStates Smart Contract.
// If no result is found, the returned IdenStateData is all zeroes.
func (ip *IdenPubOnChain) GetState(id *core.ID) (*proof.IdenStateData, error) {
	var idenState [32]byte
	var blockN uint64
	var blockTS uint64
	err := ip.client.Call(func(c *ethclient.Client) error {
		idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
		if err != nil {
			return err
		}
		blockN, blockTS, idenState, err = idenStates.GetStateDataById(nil, *id)
		return err
	})
	return &proof.IdenStateData{
		BlockN:    blockN,
		BlockTs:   int64(blockTS),
		IdenState: (*merkletree.Hash)(&idenState),
	}, err
}

// GetState returns the Identity State Data of the given ID that is closest
// (equal or older) to the queryBlockN from the IdenStates Smart Contract.  If
// a resut is found, BlockN <= queryBlockN.
// If no result is found, the returned IdenStateData is all zeroes.
func (ip *IdenPubOnChain) GetStateByBlock(id *core.ID, queryBlockN uint64) (*proof.IdenStateData, error) {
	var idenState [32]byte
	var blockN uint64
	var blockTS uint64
	err := ip.client.Call(func(c *ethclient.Client) error {
		idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
		if err != nil {
			return err
		}
		blockN, blockTS, idenState, err = idenStates.GetStateDataByBlock(nil, *id, queryBlockN)
		return err
	})
	return &proof.IdenStateData{
		BlockN:    blockN,
		BlockTs:   int64(blockTS),
		IdenState: (*merkletree.Hash)(&idenState),
	}, err
}

// GetState returns the Identity State Data of the given ID closest (equal or
// older) to the queryBlockTs from the IdenStates Smart Contract.  If a resut
// is found, BlockN <= queryBlockN.
// If no result is found, the returned IdenStateData is all zeroes.
func (ip *IdenPubOnChain) GetStateByTime(id *core.ID, queryBlockTs int64) (*proof.IdenStateData, error) {
	var idenState [32]byte
	var blockN uint64
	var blockTS uint64
	err := ip.client.Call(func(c *ethclient.Client) error {
		idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
		if err != nil {
			return err
		}
		blockN, blockTS, idenState, err = idenStates.GetStateDataByTime(nil, *id, uint64(queryBlockTs))
		return err
	})
	return &proof.IdenStateData{
		BlockN:    blockN,
		BlockTs:   int64(blockTS),
		IdenState: (*merkletree.Hash)(&idenState),
	}, err
}

// splitSignature splits the signature returning (sigR8, sigS)
func splitSignature(signature *babyjub.SignatureComp) (sigR8 [32]byte, sigS [32]byte) {
	copy(sigR8[:], signature[:32])
	copy(sigS[:], signature[32:])
	return sigR8, sigS
}

// SetState updates the Identity State of the given ID in the IdenStates Smart Contract.
func (ip *IdenPubOnChain) SetState(id *core.ID, newState *merkletree.Hash, kOpProof []byte, stateTransitionProof []byte, signature *babyjub.SignatureComp) (*types.Transaction, error) {
	if tx, err := ip.client.CallAuth(
		func(c *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
			if err != nil {
				return nil, err
			}
			sigR8, sigS := splitSignature(signature)
			return idenStates.SetState(auth, *newState, *id, kOpProof, stateTransitionProof, sigR8, sigS)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting identity state in the Smart Contract (setState): %w", err)
	} else {
		return tx, nil
	}
}

// InitState initializes the first Identity State of the given ID in the IdenStates Smart Contract.
func (ip *IdenPubOnChain) InitState(id *core.ID, genesisState *merkletree.Hash, newState *merkletree.Hash, kOpProof []byte, stateTransitionProof []byte, signature *babyjub.SignatureComp) (*types.Transaction, error) {
	if tx, err := ip.client.CallAuth(
		func(c *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
			if err != nil {
				return nil, err
			}
			sigR8, sigS := splitSignature(signature)
			return idenStates.InitState(auth, *newState, *genesisState, *id, kOpProof, stateTransitionProof, sigR8, sigS)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed initalizating identity state in the Smart Contract (initState): %w", err)
	} else {
		return tx, nil
	}
}
