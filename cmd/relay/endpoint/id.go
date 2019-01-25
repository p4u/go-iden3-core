package endpoint

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	cfg "github.com/iden3/go-iden3/cmd/relay/config"
	"github.com/iden3/go-iden3/core"
	"github.com/iden3/go-iden3/services/identitysrv"
	"github.com/iden3/go-iden3/utils"
)

// handlePostIdReq is the request used to create a new user tree in the relay.
type handlePostIdReq struct {
	//Operational   common.Address     `json:"operational"`
	OperationalPk *utils.PublicKey `json:"operationalpk" binding:"required"`
	Recoverer     common.Address   `json:"recoverer"`
	Revokator     common.Address   `json:"revokator"`
}

// handlePostIdRes is the response of a creation of a new user tree in the relay.
type handlePostIdRes struct {
	IDAddr       common.Address     `json:"idaddr"`
	ProofOfClaim *core.ProofOfClaim `json:"proofOfClaim"`
}

// handleDeployIdRes is the response of a deploy of the user contract in the blockchain.
type handleDeployIdRes struct {
	IDAddr common.Address `json:"idaddr"`
	Tx     string         `json:"tx"`
}

type handleForwardIdReq struct {
	KSignPk *utils.PublicKey `json:"ksignpk" binding:"required"`
	To      common.Address   `json:"to"`
	Data    string           `json:"data"`
	Value   string           `json:"value"`
	Gas     uint64           `json:"gas"` // gaslimit
	Sig     string           `json:"sig"`
}

type handleForwardIdRes struct {
	Tx common.Hash `json:"tx"`
}

// handleCreateId handles the creation of a new user tree from the user keys.
func handleCreateId(c *gin.Context) {

	if idservice.ImplAddr() == nil {
		fail(c, "idservice.ImplAddr()==nil", fmt.Errorf("Implementation not set"))
		return
	}

	var idreq handlePostIdReq
	if err := c.BindJSON(&idreq); err != nil {
		fail(c, "cannot parse json body", err)
		return
	}

	operational := crypto.PubkeyToAddress(idreq.OperationalPk.PublicKey)
	id := &identitysrv.Identity{
		Operational:   operational,
		OperationalPk: idreq.OperationalPk,
		Relayer:       common.HexToAddress(cfg.C.KeyStore.Address),
		Recoverer:     idreq.Recoverer,
		Revokator:     idreq.Revokator,
		Impl:          *idservice.ImplAddr(),
	}

	addr, err := idservice.AddressOf(id)
	if err != nil {
		fail(c, "failed generating identity address ", err)
		return
	}

	if proofOfClaim, err := idservice.Add(id); err != nil {
		fail(c, "failed adding identity ", err)
		return
	} else {
		c.JSON(http.StatusOK, handlePostIdRes{IDAddr: addr, ProofOfClaim: proofOfClaim})
	}
}

// handleDeployId handles the deploying of the user contract in the blockchain.
func handleDeployId(c *gin.Context) {

	idaddr := common.HexToAddress(c.Param("idaddr"))
	id, err := idservice.Get(idaddr)
	if err != nil {
		fail(c, "cannot retrieve idaddr", err)
		return
	}

	isDeployed, err := idservice.IsDeployed(idaddr)
	if err != nil {
		fail(c, "cannot retrieve deployment status", err)
		return
	}

	if isDeployed {
		fail(c, "already deployed", fmt.Errorf("already deployed"))
		return
	}

	addr, tx, err := idservice.Deploy(id)
	if err != nil {
		fail(c, "error deploying", err)
		return
	}
	c.JSON(http.StatusOK, handleDeployIdRes{addr, tx.Hash().Hex()})
}

type handleGetIdRes struct {
	IDAddr  common.Address
	LocalDb *identitysrv.Identity
	Onchain *identitysrv.Info
}

func handleGetId(c *gin.Context) {
	var idi handleGetIdRes

	idi.IDAddr = common.HexToAddress(c.Param("idaddr"))

	if info, err := idservice.Info(idi.IDAddr); err == nil {
		idi.Onchain = info
	}
	if id, err := idservice.Get(idi.IDAddr); err == nil {
		idi.LocalDb = id
	}
	c.JSON(http.StatusOK, idi)
}

func decodeBigIntParamOrFail(c *gin.Context, param, bivalue string) *big.Int {
	value := new(big.Int)
	value, ok := value.SetString(bivalue, 10)
	if !ok {
		fail(c, "bad "+param+" parameter", fmt.Errorf("bad"+param+" paremeter"))
		return nil
	}
	return value
}

func decodeHexParamOrFail(c *gin.Context, param, hexvalue string) []byte {
	if !strings.HasPrefix(hexvalue, "0x") {
		fail(c, "bad "+param+" parameter", fmt.Errorf("bad "+param+" paremeter"))
		return nil
	}
	if hexvalue == "0x0" {
		return []byte{}
	}
	data, err := hex.DecodeString(hexvalue[2:])
	if err != nil {
		fail(c, "bad data parameter", err)
		return nil
	}
	return data
}

func handleForwardId(c *gin.Context) {

	if idservice.ImplAddr() == nil {
		fail(c, "idservice.ImplAddr()==nil", fmt.Errorf("Implementation not set"))
		return
	}

	var req handleForwardIdReq
	if err := c.BindJSON(&req); err != nil {
		fail(c, "cannot parse json body", err)
		return
	}

	astxt, _ := json.MarshalIndent(req, "", "   ")
	fmt.Println(string(astxt))

	idaddr := common.HexToAddress(c.Param("idaddr"))

	var data, sig []byte
	var value *big.Int

	if data = decodeHexParamOrFail(c, "data", req.Data); data == nil {
		return
	}
	if sig = decodeHexParamOrFail(c, "sig", req.Sig); sig == nil {
		return
	}

	if value = decodeBigIntParamOrFail(c, "value", req.Value); value == nil {
		return
	}

	tx, err := idservice.Forward(idaddr,
		&req.KSignPk.PublicKey,
		req.To, data, value, req.Gas, sig)

	if err != nil {
		fail(c, "failed to forward", err)
		return
	}

	c.JSON(http.StatusOK, handleForwardIdRes{tx})
}
