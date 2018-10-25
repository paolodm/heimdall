package rest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/gorilla/mux"
	conf "github.com/maticnetwork/heimdall/helper"

	libs "github.com/maticnetwork/heimdall/libs"

	"github.com/maticnetwork/heimdall/checkpoint"
	"github.com/maticnetwork/heimdall/helper"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc(
		"/checkpoint/new",
		newCheckpointHandler(cdc, kb, cliCtx),
	).Methods("POST")
}

type EpochCheckpoint struct {
	RootHash        string `json:"root_hash"`
	StartBlock      int64  `json:"start_block"`
	EndBlock        int64  `json:"end_block"`
	ProposerAddress string `json:"proposer_address"`
}

func newCheckpointHandler(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	logger := conf.Logger.With("module", "checkpoint/rest/tx")

	return func(w http.ResponseWriter, r *http.Request) {
		var m EpochCheckpoint

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		err = json.Unmarshal(body, &m)
		if err != nil {
			logger.Error("Error Unmarshalling json Epoch Checkpoint", "Error", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		msg := checkpoint.NewMsgCheckpointBlock(
			uint64(m.StartBlock),
			uint64(m.EndBlock),
			common.HexToHash(m.RootHash),
			m.ProposerAddress,
		)

		tx := checkpoint.NewBaseTx(msg)

		txBytes, err := rlp.EncodeToBytes(tx)
		if err != nil {
			logger.Error("Error generating TX Bytes ", "Error", err)
		}
		logger.Info("The tx bytes are ", "Transaction: ", hex.EncodeToString(txBytes))

		logger.Info("URL to broadcast tx ", "URL", helper.GetConfig().TendermintEndpoint)

		resp := sendRequest(txBytes, helper.GetConfig().TendermintEndpoint, logger)
		logger.Info("Transaction Sent !  ", "Response ", resp)

		var bodyString string
		if resp.StatusCode == http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			bodyString = string(bodyBytes)
		}
		w.Write([]byte(bodyString))
	}
}

func sendRequest(txBytes []byte, url string, logger libs.Logger) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/broadcast_tx_commit", nil)
	if err != nil {
		logger.Error("Error while drafting request for tendermint", "Error", err)
	}

	queryParams := req.URL.Query()
	queryParams.Add("tx", fmt.Sprintf("0x%s", hex.EncodeToString(txBytes)))
	req.URL.RawQuery = queryParams.Encode()

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error while sending request to tendermint", "Error", err)
	}
	return resp
}
