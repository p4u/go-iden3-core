package idenpuboffchainwriter

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/iden3/go-iden3-core/db"
	"github.com/iden3/go-iden3-core/merkletree"
)

var (
	ErrIdenStateNotFound = fmt.Errorf("identity state not found in the cache")
)

var (
	dbKeyConfig          = []byte("config")
	dbKeyCacheIdx        = []byte("cacheidx")
	dbKeyIdenState       = []byte("idenstate")
	dbKeyClaimsRoot      = []byte("claimsroot")
	dbKeyRootsRoot       = []byte("rootsroot")
	dbKeyRevocationsRoot = []byte("revocationsroot")
	dbKeyRootsTree       = []byte("rootstree")
	dbKeyRevocationsTree = []byte("revocationstree")
)

// IdenPubOffChainWriter is a interface to write the off chain public state of an identity.
type IdenPubOffChainWriter interface {
	Publish(idenState, claimsRoot, revocationsRoot, rootsRoot *merkletree.Hash) error
}

var ConfigDefault = Config{CacheLen: 1}

type Config struct {
	CacheLen byte
}

// IdenPubOffChainWriteHttp satisfies the IdenPubOffChainWriter interface, and stores in a leveldb the published RootsTree & RevocationsTree to be returned when requested.
type IdenPubOffChainWriteHttp struct {
	rw              *sync.RWMutex
	storage         db.Storage
	rootsTree       *merkletree.MerkleTree
	revocationsTree *merkletree.MerkleTree
	cfg             *Config
}

// NewIdenPubOffChainWriteHttp returns a new IdenPubOffChainWriteHttp
func NewIdenPubOffChainWriteHttp(cfg *Config, storage db.Storage, rootsTree *merkletree.MerkleTree, revocationsTree *merkletree.MerkleTree) (*IdenPubOffChainWriteHttp, error) {
	i := IdenPubOffChainWriteHttp{
		rw:              &sync.RWMutex{},
		storage:         storage,
		rootsTree:       rootsTree,
		revocationsTree: revocationsTree,
		cfg:             cfg,
	}
	tx, err := i.storage.NewTx()
	if err != nil {
		return nil, err
	}
	i.initCacheIdx(tx)
	if err := db.StoreJSON(tx, dbKeyConfig, &cfg); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &i, nil
}

// LoadIdenPubOffChainWriteHttp returns a new IdenPubOffChainWriteHttp
func LoadIdenPubOffChainWriteHttp(storage db.Storage, rootsTree *merkletree.MerkleTree, revocationsTree *merkletree.MerkleTree) (*IdenPubOffChainWriteHttp, error) {
	var cfg Config
	if err := db.LoadJSON(storage, dbKeyConfig, &cfg); err != nil {
		return nil, err
	}
	i := IdenPubOffChainWriteHttp{
		rw:              &sync.RWMutex{},
		storage:         storage,
		rootsTree:       rootsTree,
		revocationsTree: revocationsTree,
		cfg:             &cfg,
	}
	return &i, nil
}

// Publish publishes the RootsTree and RevocationsTree to the configured way of publishing
func (i *IdenPubOffChainWriteHttp) Publish(idenState, claimsRoot, revocationsRoot, rootsRoot *merkletree.Hash) error {
	// RootsTree
	w := bytes.NewBufferString("")
	err := i.rootsTree.DumpTree(w, rootsRoot)
	if err != nil {
		return err
	}
	rotBlob := w.Bytes()

	// RevocationsTree
	w = bytes.NewBufferString("")
	err = i.revocationsTree.DumpTree(w, revocationsRoot)
	if err != nil {
		return err
	}
	retBlob := w.Bytes()

	tx, err := i.storage.NewTx()
	if err != nil {
		return err
	}
	i.rw.Lock()
	defer func() {
		if err == nil {
			if err := tx.Commit(); err != nil {
				tx.Close()
			}
		} else {
			tx.Close()
		}
		i.rw.Unlock()
	}()

	cacheIdx, err := i.nextCacheIdx(tx)
	if err != nil {
		return err
	}

	tx.Put(append(dbKeyIdenState, cacheIdx), idenState[:])
	tx.Put(append(dbKeyClaimsRoot, cacheIdx), claimsRoot[:])
	tx.Put(append(dbKeyRootsRoot, cacheIdx), rootsRoot[:])
	tx.Put(append(dbKeyRootsTree, cacheIdx), rotBlob)
	tx.Put(append(dbKeyRevocationsRoot, cacheIdx), revocationsRoot[:])
	tx.Put(append(dbKeyRevocationsTree, cacheIdx), retBlob)

	return nil
}

func (i *IdenPubOffChainWriteHttp) prevCacheIdx(tx db.Tx) (byte, error) {
	cacheIdx, err := tx.Get(dbKeyCacheIdx)
	if err != nil {
		return 0, err
	}
	return (cacheIdx[0] - 1) % i.cfg.CacheLen, nil
}

// nextCacheIdx returns the current cacheIdx and stores the next one.
func (i *IdenPubOffChainWriteHttp) nextCacheIdx(tx db.Tx) (byte, error) {
	cacheIdx, err := tx.Get(dbKeyCacheIdx)
	if err != nil {
		return 0, err
	}
	next := (cacheIdx[0] + 1) % i.cfg.CacheLen
	tx.Put(dbKeyCacheIdx, []byte{next})
	return cacheIdx[0], nil
}

func (i *IdenPubOffChainWriteHttp) initCacheIdx(tx db.Tx) {
	tx.Put(dbKeyCacheIdx, []byte{0})
}

// func (i *IdenPubOffChainWriteHttp) getCacheIdx(tx db.Tx) (byte, error) {
// 	cacheIdx, err := tx.Get(dbKeyCacheIdx)
// 	if err == db.ErrNotFound {
// 		cacheIdx = []byte{0}
// 	} else if err != nil {
// 		return 0, err
// 	}
// 	return cacheIdx[0], nil
// }

// PublicData contains the RootsTree + Root, and the RevocationTree + Root
type PublicData struct {
	IdenState           merkletree.Hash
	ClaimsTreeRoot      merkletree.Hash
	RootsTreeRoot       merkletree.Hash
	RootsTree           []byte
	RevocationsTreeRoot merkletree.Hash
	RevocationsTree     []byte
}

// GetPublicData returns the identity off chain public data corresponding to
// the queryIdenState.  If the queryIdenState is nil, the last identity off
// chain public data is returned.
func (i *IdenPubOffChainWriteHttp) GetPublicData(queryIdenState *merkletree.Hash) (*PublicData, error) {
	tx, err := i.storage.NewTx()
	if err != nil {
		return nil, err
	}
	defer tx.Close()
	i.rw.RLock()
	defer i.rw.RUnlock()

	var cacheIdx byte
	if queryIdenState == nil {
		cacheIdx, err = i.prevCacheIdx(tx)
		if err != nil {
			return nil, err
		}
	} else {
		idx := byte(0)
		for ; idx < i.cfg.CacheLen; idx++ {
			// idenState
			idenState, err := tx.Get(append(dbKeyIdenState, idx))
			if err != nil {
				return nil, err
			}
			if bytes.Equal(queryIdenState[:], idenState) {
				break
			}
		}
		if idx == i.cfg.CacheLen {
			return nil, ErrIdenStateNotFound
		}
	}
	// idenState
	idenState, err := tx.Get(append(dbKeyIdenState, cacheIdx))
	if err != nil {
		return nil, err
	}

	// claims tree root
	cltRoot, err := tx.Get(append(dbKeyClaimsRoot, cacheIdx))
	if err != nil {
		return nil, err
	}

	// roots tree
	rotRoot, err := tx.Get(append(dbKeyRootsRoot, cacheIdx))
	if err != nil {
		return nil, err
	}
	rot, err := tx.Get(append(dbKeyRootsTree, cacheIdx))
	if err != nil {
		return nil, err
	}

	// revocations tree
	retRoot, err := tx.Get(append(dbKeyRevocationsRoot, cacheIdx))
	if err != nil {
		return nil, err
	}
	ret, err := tx.Get(append(dbKeyRevocationsTree, cacheIdx))
	if err != nil {
		return nil, err
	}

	var idenState32 [merkletree.ElemBytesLen]byte
	var cltRoot32 [merkletree.ElemBytesLen]byte
	var rotRoot32 [merkletree.ElemBytesLen]byte
	var retRoot32 [merkletree.ElemBytesLen]byte
	copy(idenState32[:], idenState[:32])
	copy(cltRoot32[:], cltRoot[:32])
	copy(rotRoot32[:], rotRoot[:32])
	copy(retRoot32[:], retRoot[:32])

	p := &PublicData{
		IdenState:           merkletree.Hash(merkletree.ElemBytes(idenState32)),
		ClaimsTreeRoot:      merkletree.Hash(merkletree.ElemBytes(cltRoot32)),
		RootsTreeRoot:       merkletree.Hash(merkletree.ElemBytes(rotRoot32)),
		RootsTree:           rot,
		RevocationsTreeRoot: merkletree.Hash(merkletree.ElemBytes(retRoot32)),
		RevocationsTree:     ret,
	}
	return p, nil
}
