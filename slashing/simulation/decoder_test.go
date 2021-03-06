package simulation

import (
	"fmt"
	"testing"
	"time"

	gogotypes "github.com/gogo/protobuf/types"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/slashing/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// nolint:deadcode,unused,varcheck
var (
	delPk1    = secp256k1.GenPrivKey().PubKey()
	delAddr1  = sdk.AccAddress(delPk1.Address())
	valAddr1  = sdk.ValAddress(delPk1.Address())
	consAddr1 = hmTypes.BytesToHeimdallAddress(delPk1.Address().Bytes())
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	types.RegisterCodec(cdc)
	return
}

func TestDecodeStore(t *testing.T) {
	cdc := makeTestCodec()

	info := hmTypes.NewValidatorSigningInfo(consAddr1, 0, 1, time.Now().UTC(), false, 0)
	bechPK := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, delPk1)
	missed := gogotypes.BoolValue{Value: true}

	kvPairs := tmkv.Pairs{
		tmkv.Pair{Key: types.GetValidatorSigningInfoKey(consAddr1), Value: cdc.MustMarshalBinaryBare(info)},
		tmkv.Pair{Key: types.GetValidatorMissedBlockBitArrayKey(consAddr1, 6), Value: cdc.MustMarshalBinaryBare(&missed)},
		tmkv.Pair{Key: types.GetAddrPubkeyRelationKey(delAddr1), Value: cdc.MustMarshalBinaryBare(delPk1)},
		tmkv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"ValidatorSigningInfo", fmt.Sprintf("%v\n%v", info, info)},
		{"ValidatorMissedBlockBitArray", fmt.Sprintf("missedA: %v\nmissedB: %v", missed.Value, missed.Value)},
		{"AddrPubkeyRelation", fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", bechPK, bechPK)},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeStore(cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeStore(cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
