package client

import (
	"context"
	"errors"
	"reflect"
	"testing"

	sdkRpc "github.com/blocto/solana-go-sdk/rpc"
	"github.com/skport/solana-rpc-client-extensions-go/types"
)

func TestClient_GetStakeActivation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		address string
		want    *GetStakeActivationResponse
		wantErr error
	}{
		// Stake Accounts
		// https://explorer.solana.com/address/HmbKSyhneFd1Nd8BtW7ejBHTFrbnBsVA7JE6GpA9WjiX?cluster=devnet
		// {
		// 	name:    "Testnet Success 01: inactive stake",
		// 	address: "HmbKSyhneFd1Nd8BtW7ejBHTFrbnBsVA7JE6GpA9WjiX",
		// 	want: &GetStakeActivationResponse{
		// 		Active:   0,
		// 		Inactive: 950000000,  // lamports = 0.952282880 SOL
		// 		State:    "inactive", // active or inactive or activating or deactivating
		// 	},
		// 	wantErr: nil,
		// },
		// {
		// 	name:    "Testnet Success 02: activating stake",
		// 	address: "HmbKSyhneFd1Nd8BtW7ejBHTFrbnBsVA7JE6GpA9WjiX",
		// 	want: &GetStakeActivationResponse{
		// 		Active:   0,
		// 		Inactive: 950000000,
		// 		State:    "activating",
		// 	},
		// 	wantErr: nil,
		// },
		// {
		// 	name:    "Testnet Success 03: active stake",
		// 	address: "HmbKSyhneFd1Nd8BtW7ejBHTFrbnBsVA7JE6GpA9WjiX",
		// 	want: &GetStakeActivationResponse{
		// 		Active:   950000000,
		// 		Inactive: 0,
		// 		State:    "active",
		// 	},
		// 	wantErr: nil,
		// },
		// {
		// 	name:    "Testnet Success 04: deavtivating stake",
		// 	address: "HmbKSyhneFd1Nd8BtW7ejBHTFrbnBsVA7JE6GpA9WjiX",
		// 	want: &GetStakeActivationResponse{
		// 		Active:   0,
		// 		Inactive: 950000000,
		// 		State:    "deactivating",
		// 	},
		// 	wantErr: nil,
		// },
	}

	// GetAccountInfo for stakeHistoryAccount
	stakeHistoryAccountInfo, err := rpc.GetAccountInfoWithConfig(ctx, StakeHistoryAccountAddress,
		sdkRpc.GetAccountInfoConfig{
			Commitment: sdkRpc.CommitmentFinalized,
			Encoding:   sdkRpc.AccountEncodingJsonParsed,
		},
	)
	if err != nil {
		t.Errorf("GetAccountInfo error for stakeHistoryAccount: %v", err)
		return
	}
	stakeHistoryAccount, err := ConvertStakeHistoryAccountInfo(stakeHistoryAccountInfo)
	if err != nil {
		t.Errorf("ConvertStakeHistoryAccountInfo error: %v", err)
		return
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GetEpochInfo
			epochInfo, err := rpc.GetEpochInfo(ctx)
			if err != nil {
				t.Errorf("GetEpochInfo error: %v", err)
				return
			}

			// GetAccountInfo for stakeAccount
			stakeAccountInfo, err := rpc.GetAccountInfoWithConfig(ctx, tt.address,
				sdkRpc.GetAccountInfoConfig{
					Commitment: sdkRpc.CommitmentFinalized,
					Encoding:   sdkRpc.AccountEncodingJsonParsed,
				},
			)
			if err != nil {
				t.Errorf("GetAccountInfo error for stakeAccount: %v", err)
				return
			}
			stakeAccount, err := ConvertStakeAccountInfo(stakeAccountInfo)
			if err != nil {
				t.Errorf("ConvertStakeAccountInfo error: %v", err)
				return
			}

			r, err := GetStakeActivation(tt.address, epochInfo.Result.Epoch, stakeAccount, stakeHistoryAccount)
			if err != nil {
				t.Errorf("GetStakeActivation error: %v", err)
			} else {
				t.Logf("GetStakeActivation: %v", r)
			}

			if !reflect.DeepEqual(tt.want, r) {
				t.Errorf("GetStakeActivation = %v, want %v", r, tt.want)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetStakeActivation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getSolanaStakeHistoryEntry(t *testing.T) {
	tests := []struct {
		name       string
		historyAct *types.StakeHistoryAccount
		epoch      uint64
		want       *types.StakeHistoryAccountInfo
	}{
		{
			name: "success 1",
			historyAct: &types.StakeHistoryAccount{
				Data: struct {
					Parsed struct {
						Info []types.StakeHistoryAccountInfo `json:"info"`
						Type string                          `json:"type"`
					} `json:"parsed"`
					Program string `json:"program"`
					Space   int    `json:"space"`
				}{
					Parsed: struct {
						Info []types.StakeHistoryAccountInfo `json:"info"`
						Type string                          `json:"type"`
					}{
						Info: []types.StakeHistoryAccountInfo{
							{
								Epoch: 1,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   100,
									Deactivating: 10,
									Effective:    90,
								},
							},
							{
								Epoch: 100,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   200,
									Deactivating: 50,
									Effective:    150,
								},
							},
						},
						Type: "stakeHistory",
					},
					Program: "stakeHistoryProgram",
					Space:   2048,
				},
				Executable: false,
				Lamports:   0,
				Owner:      "",
				RentEpoch:  0,
				Space:      0,
			},
			epoch: 100,
			want: &types.StakeHistoryAccountInfo{
				Epoch: 100,
				StakeHistory: struct {
					Activating   uint64 `json:"activating"`
					Deactivating uint64 `json:"deactivating"`
					Effective    uint64 `json:"effective"`
				}{
					Activating:   200,
					Deactivating: 50,
					Effective:    150,
				},
			},
		},
		{
			name: "nil: 1",
			historyAct: &types.StakeHistoryAccount{
				Data: struct {
					Parsed struct {
						Info []types.StakeHistoryAccountInfo `json:"info"`
						Type string                          `json:"type"`
					} `json:"parsed"`
					Program string `json:"program"`
					Space   int    `json:"space"`
				}{
					Parsed: struct {
						Info []types.StakeHistoryAccountInfo `json:"info"`
						Type string                          `json:"type"`
					}{
						Info: []types.StakeHistoryAccountInfo{
							{
								Epoch: 1,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   100,
									Deactivating: 10,
									Effective:    90,
								},
							},
						},
						Type: "stakeHistory",
					},
					Program: "stakeHistoryProgram",
					Space:   2048,
				},
				Executable: false,
				Lamports:   0,
				Owner:      "",
				RentEpoch:  0,
				Space:      0,
			},
			epoch: 200,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := getSolanaStakeHistoryEntry(tt.historyAct, tt.epoch)
			if !reflect.DeepEqual(tt.want, r) {
				t.Errorf("getSolanaStakeHistoryEntry = %v, want %v", r, tt.want)
			}
		})
	}
}
