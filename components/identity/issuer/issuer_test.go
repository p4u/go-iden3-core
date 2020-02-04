package issuer

import (
	"testing"

	"github.com/iden3/go-iden3-core/db"
	"github.com/iden3/go-iden3-core/keystore"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/stretchr/testify/require"
)

var pass = []byte("my passphrase")

func TestNewLoadIssuer(t *testing.T) {
	cfg := ConfigDefault
	storage := db.NewMemoryStorage()
	ksStorage := keystore.MemStorage([]byte{})
	keyStore, err := keystore.NewKeyStore(&ksStorage, keystore.LightKeyStoreParams)
	require.Nil(t, err)
	kOp, err := keyStore.NewKey(pass)
	require.Nil(t, err)
	issuer, err := New(cfg, kOp, []merkletree.Entrier{}, storage, keyStore, nil)
	require.Nil(t, err)

	issuerLoad, err := Load(storage, keyStore, nil)
	require.Nil(t, err)

	require.Equal(t, issuer.cfg, issuerLoad.cfg)
	require.Equal(t, issuer.id, issuerLoad.id)
}