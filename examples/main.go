package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	sdkRpc "github.com/blocto/solana-go-sdk/rpc"
	"github.com/skport/solana-rpc-client-extensions-go/client"
)

const (
	stakeAccountAddress = "HmbKSyhneFd1Nd8BtW7ejBHTFrbnBsVA7JE6GpA9WjiX"
)

func main() {
	rpc := sdkRpc.NewRpcClient(sdkRpc.DevnetRPCEndpoint)
	ctx := context.Background()

	// GetEpochInfo
	epochInfo, err := rpc.GetEpochInfo(ctx)
	if err != nil {
		log.Panicf("GetEpochInfo error: %v", err)
	}

	// GetAccountInfo for stakeHistoryAccount
	stakeHistoryAccountInfo, err := rpc.GetAccountInfoWithConfig(ctx, client.StakeHistoryAccountAddress,
		sdkRpc.GetAccountInfoConfig{
			Commitment: sdkRpc.CommitmentFinalized,
			Encoding:   sdkRpc.AccountEncodingJsonParsed,
		},
	)
	if err != nil {
		log.Panicf("GetAccountInfo error for stakeHistoryAccount: %v", err)
	}
	stakeHistoryAccount, err := client.ConvertStakeHistoryAccountInfo(stakeHistoryAccountInfo)
	if err != nil {
		log.Panicf("ConvertStakeHistoryAccountInfo error: %v", err)
	}

	// GetAccountInfo for stakeAccount
	stakeAccountInfo, err := rpc.GetAccountInfoWithConfig(ctx, stakeAccountAddress,
		sdkRpc.GetAccountInfoConfig{
			Commitment: sdkRpc.CommitmentFinalized,
			Encoding:   sdkRpc.AccountEncodingJsonParsed,
		},
	)
	if err != nil {
		log.Panicf("GetAccountInfo error for stakeAccount: %v", err)
	}
	stakeAccount, err := client.ConvertStakeAccountInfo(stakeAccountInfo)
	if err != nil {
		log.Panicf("ConvertStakeAccountInfo error: %v", err)
	}

	// GetStakeActivation
	r, err := client.GetStakeActivation(stakeAccountAddress, epochInfo.Result.Epoch, stakeAccount, stakeHistoryAccount)
	if err != nil {
		log.Panicf("GetStakeActivation error: %v", err)
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Panicf("failed to marshal JSON: %v", err)
	}
	fmt.Printf("GetStakeActivation: %s", string(b))
}
