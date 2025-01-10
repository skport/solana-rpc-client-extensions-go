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

func Test_getSolanaStakeActivatingAndDeactivating(t *testing.T) {
	tests := []struct {
		name string

		stakeAccountAddress string
		stakeAccount        *types.StakeAccount

		targetEpoch         uint64
		stakeHistoryAccount *types.StakeHistoryAccount // Official account that maintains staking history (returns the amount of staking for the entire network per epoch)

		// Expected statuses
		wantEffective    uint64 // effective amount of staking
		wantActivating   uint64 // staking amount while activating
		wantDeactivating uint64 // staking amount while deactivating
		wantErr          error
	}{
		// ─────────────────────────────────────────────
		// Devnet
		// ─────────────────────────────────────────────
		{
			name:                "success 1: activating",
			stakeAccountAddress: "55pRDNDdQBNWfFRQy7eDSz2yyLs5n8ckbTGrtnD5miaQ",
			stakeAccount: &types.StakeAccount{
				Data: struct {
					Parsed struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					} `json:"parsed"`
					Program string `json:"program"`
					Space   uint64 `json:"space"`
				}{
					Parsed: struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					}{
						Info: struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						}{
							Meta: types.StakeAccountInfoMeta{
								Authorized: struct {
									Staker     string `json:"staker"`
									Withdrawer string `json:"withdrawer"`
								}{
									Staker:     "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
									Withdrawer: "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
								},
								Lockup: struct {
									Custodian     string `json:"custodian"`
									Epoch         uint64 `json:"epoch"`
									UnixTimestamp uint64 `json:"unixTimestamp"`
								}{
									Custodian:     "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
									Epoch:         0,
									UnixTimestamp: 0,
								},
								RentExemptReserve: "2282880",
							},
							Stake: &types.StakeAccountInfoStake{
								CreditsObserved: 612480517,
								Delegation: struct {
									ActivationEpoch    string  `json:"activationEpoch"`
									DeactivationEpoch  string  `json:"deactivationEpoch"`
									Stake              string  `json:"stake"`
									Voter              string  `json:"voter"`
									WarmupCooldownRate float64 `json:"warmupCooldownRate"`
								}{
									ActivationEpoch:    "816", // When the epoch at the time of delegation is 817, it becomes active.
									DeactivationEpoch:  "18446744073709551615",
									Stake:              "1000000000",
									Voter:              "FwR3PbjS5iyqzLiLugrBqKSa5EKZ4vK9SKs7eQXtT59f",
									WarmupCooldownRate: 0.25,
								},
							},
						},
						Type: "delegated",
					},
					Program: "stake",
					Space:   200,
				},
				Executable: false,
				Lamports:   1002282880,
				Owners:     "Stake11111111111111111111111111111111111111",
				RentEpoch:  1844674407370955200,
			},
			targetEpoch: 816,
			stakeHistoryAccount: &types.StakeHistoryAccount{
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
						// In Epoch 816, you can get the history up to 815, but not 816 (because the Epoch has not been completed).
						Info: []types.StakeHistoryAccountInfo{
							{
								Epoch: 815,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
							{
								Epoch: 814,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
							{
								Epoch: 813,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
						},
						Type: "stakeHistory",
					},
					Program: "stakeHistoryProgram",
					Space:   2048,
				},
			},
			wantEffective:    0,
			wantActivating:   1000000000,
			wantDeactivating: 0,
			wantErr:          nil,
		},
		{
			name:                "success 2: active",
			stakeAccountAddress: "55pRDNDdQBNWfFRQy7eDSz2yyLs5n8ckbTGrtnD5miaQ",
			stakeAccount: &types.StakeAccount{
				Data: struct {
					Parsed struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					} `json:"parsed"`
					Program string `json:"program"`
					Space   uint64 `json:"space"`
				}{
					Parsed: struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					}{
						Info: struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						}{
							Meta: types.StakeAccountInfoMeta{
								Authorized: struct {
									Staker     string `json:"staker"`
									Withdrawer string `json:"withdrawer"`
								}{
									Staker:     "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
									Withdrawer: "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
								},
								Lockup: struct {
									Custodian     string `json:"custodian"`
									Epoch         uint64 `json:"epoch"`
									UnixTimestamp uint64 `json:"unixTimestamp"`
								}{
									Custodian:     "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
									Epoch:         0,
									UnixTimestamp: 0,
								},
								RentExemptReserve: "2282880",
							},
							Stake: &types.StakeAccountInfoStake{
								CreditsObserved: 612480517,
								Delegation: struct {
									ActivationEpoch    string  `json:"activationEpoch"`
									DeactivationEpoch  string  `json:"deactivationEpoch"`
									Stake              string  `json:"stake"`
									Voter              string  `json:"voter"`
									WarmupCooldownRate float64 `json:"warmupCooldownRate"`
								}{
									ActivationEpoch:    "816", // When the delegated epoch reaches 817, it becomes active.
									DeactivationEpoch:  "18446744073709551615",
									Stake:              "1000000000",
									Voter:              "FwR3PbjS5iyqzLiLugrBqKSa5EKZ4vK9SKs7eQXtT59f",
									WarmupCooldownRate: 0.25,
								},
							},
						},
						Type: "delegated",
					},
					Program: "stake",
					Space:   200,
				},
				Executable: false,
				Lamports:   1002282880,
				Owners:     "Stake11111111111111111111111111111111111111",
				RentEpoch:  1844674407370955200,
			},
			targetEpoch: 817,
			stakeHistoryAccount: &types.StakeHistoryAccount{
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
						// In Epoch 817, you can get the history up to 816, but not 817.
						Info: []types.StakeHistoryAccountInfo{
							{
								Epoch: 816,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
							{
								Epoch: 815,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
							{
								Epoch: 814,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
						},
						Type: "stakeHistory",
					},
					Program: "stakeHistoryProgram",
					Space:   2048,
				},
			},
			wantEffective:    1000000000, // When it reaches 817, the staking that was being activated at 816 will become active.
			wantActivating:   0,
			wantDeactivating: 0,
			wantErr:          nil,
		},
		{
			name:                "success 3: deactivating 1",
			stakeAccountAddress: "55pRDNDdQBNWfFRQy7eDSz2yyLs5n8ckbTGrtnD5miaQ",
			stakeAccount: &types.StakeAccount{
				Data: struct {
					Parsed struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					} `json:"parsed"`
					Program string `json:"program"`
					Space   uint64 `json:"space"`
				}{
					Parsed: struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					}{
						Info: struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						}{
							Meta: types.StakeAccountInfoMeta{
								Authorized: struct {
									Staker     string `json:"staker"`
									Withdrawer string `json:"withdrawer"`
								}{
									Staker:     "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
									Withdrawer: "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
								},
								Lockup: struct {
									Custodian     string `json:"custodian"`
									Epoch         uint64 `json:"epoch"`
									UnixTimestamp uint64 `json:"unixTimestamp"`
								}{
									Custodian:     "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
									Epoch:         0,
									UnixTimestamp: 0,
								},
								RentExemptReserve: "2282880",
							},
							Stake: &types.StakeAccountInfoStake{
								CreditsObserved: 612480517,
								Delegation: struct {
									ActivationEpoch    string  `json:"activationEpoch"`
									DeactivationEpoch  string  `json:"deactivationEpoch"`
									Stake              string  `json:"stake"`
									Voter              string  `json:"voter"`
									WarmupCooldownRate float64 `json:"warmupCooldownRate"`
								}{
									ActivationEpoch:    "816",
									DeactivationEpoch:  "817", // When the epoch at the time of deactivation is 818, it becomes inactive.
									Stake:              "1000000000",
									Voter:              "FwR3PbjS5iyqzLiLugrBqKSa5EKZ4vK9SKs7eQXtT59f",
									WarmupCooldownRate: 0.25,
								},
							},
						},
						Type: "delegated",
					},
					Program: "stake",
					Space:   200,
				},
				Executable: false,
				Lamports:   1002282880,
				Owners:     "Stake11111111111111111111111111111111111111",
				RentEpoch:  1844674407370955200,
			},
			targetEpoch: 817,
			stakeHistoryAccount: &types.StakeHistoryAccount{
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
						// In Epoch 817, you can get the history up to 816, but not 817.
						Info: []types.StakeHistoryAccountInfo{
							{
								Epoch: 816,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
							{
								Epoch: 815,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
							{
								Epoch: 814,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
						},
						Type: "stakeHistory",
					},
					Program: "stakeHistoryProgram",
					Space:   2048,
				},
			},
			wantEffective:    1000000000, // 817 is being deactivated, so stakes are still valid.
			wantActivating:   0,
			wantDeactivating: 1000000000, // When it reaches 818, the staking that was being invalidated at 817 will be invalidated.
			wantErr:          nil,
		},
		{
			name:                "success 3: deactivating 2",
			stakeAccountAddress: "55pRDNDdQBNWfFRQy7eDSz2yyLs5n8ckbTGrtnD5miaQ",
			stakeAccount: &types.StakeAccount{
				Data: struct {
					Parsed struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					} `json:"parsed"`
					Program string `json:"program"`
					Space   uint64 `json:"space"`
				}{
					Parsed: struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					}{
						Info: struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						}{
							Meta: types.StakeAccountInfoMeta{
								Authorized: struct {
									Staker     string `json:"staker"`
									Withdrawer string `json:"withdrawer"`
								}{
									Staker:     "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
									Withdrawer: "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
								},
								Lockup: struct {
									Custodian     string `json:"custodian"`
									Epoch         uint64 `json:"epoch"`
									UnixTimestamp uint64 `json:"unixTimestamp"`
								}{
									Custodian:     "3oexKwZRXJNwJjaaLCrqYVMauS4EQAk7zzhScuqTQD77",
									Epoch:         0,
									UnixTimestamp: 0,
								},
								RentExemptReserve: "2282880",
							},
							Stake: &types.StakeAccountInfoStake{
								CreditsObserved: 612480517,
								Delegation: struct {
									ActivationEpoch    string  `json:"activationEpoch"`
									DeactivationEpoch  string  `json:"deactivationEpoch"`
									Stake              string  `json:"stake"`
									Voter              string  `json:"voter"`
									WarmupCooldownRate float64 `json:"warmupCooldownRate"`
								}{
									ActivationEpoch:    "816",
									DeactivationEpoch:  "818", // When the epoch at the time of deactivation is 819, it becomes inactive.
									Stake:              "1000000000",
									Voter:              "FwR3PbjS5iyqzLiLugrBqKSa5EKZ4vK9SKs7eQXtT59f",
									WarmupCooldownRate: 0.25,
								},
							},
						},
						Type: "delegated",
					},
					Program: "stake",
					Space:   200,
				},
				Executable: false,
				Lamports:   1002282880,
				Owners:     "Stake11111111111111111111111111111111111111",
				RentEpoch:  1844674407370955200,
			},
			targetEpoch: 818,
			stakeHistoryAccount: &types.StakeHistoryAccount{
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
						// Epoch 818 can record up to 817, but not 818.
						Info: []types.StakeHistoryAccountInfo{
							{
								Epoch: 817,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
							{
								Epoch: 816,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
							{
								Epoch: 815,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1847742172715,
									Deactivating: 5465100758,
									Effective:    169798767116673467,
								},
							},
						},
						Type: "stakeHistory",
					},
					Program: "stakeHistoryProgram",
					Space:   2048,
				},
			},
			wantEffective:    1000000000, // 818 is still being deactivated, so it is valid.
			wantActivating:   0,
			wantDeactivating: 1000000000, // When it reaches 819, the deactivation is complete.
			wantErr:          nil,
		},

		// ─────────────────────────────────────────────
		// Mainnet
		// ─────────────────────────────────────────────
		{
			name:                "success 1 (Mainnet): active",
			stakeAccountAddress: "55pRDNDdQBNWfFRQy7eDSz2yyLs5n8ckbTGrtnD5miaQ",
			stakeAccount: &types.StakeAccount{
				Data: struct {
					Parsed struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					} `json:"parsed"`
					Program string `json:"program"`
					Space   uint64 `json:"space"`
				}{
					Parsed: struct {
						Info struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						} `json:"info"`
						Type string `json:"type"`
					}{
						Info: struct {
							Meta  types.StakeAccountInfoMeta   `json:"meta"`
							Stake *types.StakeAccountInfoStake `json:"stake"`
						}{
							Meta: types.StakeAccountInfoMeta{
								Authorized: struct {
									Staker     string `json:"staker"`
									Withdrawer string `json:"withdrawer"`
								}{
									Staker:     "DykQKrexKLfA7eUUSQffi2hmzrB1UEYoY2BpesHH4jcd",
									Withdrawer: "DykQKrexKLfA7eUUSQffi2hmzrB1UEYoY2BpesHH4jcd",
								},
								Lockup: struct {
									Custodian     string `json:"custodian"`
									Epoch         uint64 `json:"epoch"`
									UnixTimestamp uint64 `json:"unixTimestamp"`
								}{
									Custodian:     "11111111111111111111111111111111",
									Epoch:         0,
									UnixTimestamp: 0,
								},
								RentExemptReserve: "2282880",
							},
							Stake: &types.StakeAccountInfoStake{
								CreditsObserved: 612480517,
								Delegation: struct {
									ActivationEpoch    string  `json:"activationEpoch"`
									DeactivationEpoch  string  `json:"deactivationEpoch"`
									Stake              string  `json:"stake"`
									Voter              string  `json:"voter"`
									WarmupCooldownRate float64 `json:"warmupCooldownRate"`
								}{
									ActivationEpoch:    "694", // Epoch at the time of delegation
									DeactivationEpoch:  "18446744073709551615",
									Stake:              "7799841",
									Voter:              "J2nUHEAgZFRyuJbFjdqPrAa9gyWDuc7hErtDQHPhsYRp",
									WarmupCooldownRate: 0.25,
								},
							},
						},
						Type: "delegated",
					},
					Program: "stake",
					Space:   200,
				},
				Executable: false,
				Lamports:   10114552,
				Owners:     "Stake11111111111111111111111111111111111111",
				RentEpoch:  1844674407370955200,
			},
			targetEpoch: 724,
			stakeHistoryAccount: &types.StakeHistoryAccount{
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
								Epoch: 723,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   4187642636325805,
									Deactivating: 2967867904941421,
									Effective:    389840397775808821,
								},
							},
							{
								Epoch: 722,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   2883390354002212,
									Deactivating: 1971880643345629,
									Effective:    388817334675841644,
								},
							},
							{
								Epoch: 721,
								StakeHistory: struct {
									Activating   uint64 `json:"activating"`
									Deactivating uint64 `json:"deactivating"`
									Effective    uint64 `json:"effective"`
								}{
									Activating:   1778721828311106,
									Deactivating: 2701448216401765,
									Effective:    389628586371016017,
								},
							},
						},
						Type: "stakeHistory",
					},
					Program: "stakeHistoryProgram",
					Space:   2048,
				},
			},
			wantEffective:    7799841,
			wantActivating:   0,
			wantDeactivating: 0,
			wantErr:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := getSolanaStakeActivatingAndDeactivating(tt.stakeAccountAddress, tt.stakeAccount, tt.targetEpoch, tt.stakeHistoryAccount)

			if !reflect.DeepEqual(tt.wantEffective, got) {
				t.Errorf("wantEffective = %v, want %v", got, tt.wantEffective)
			}
			if !reflect.DeepEqual(tt.wantActivating, got1) {
				t.Errorf("wantActivating = %v, want %v", got1, tt.wantActivating)
			}
			if !reflect.DeepEqual(tt.wantDeactivating, got2) {
				t.Errorf("wantDeactivating = %v, want %v", got2, tt.wantDeactivating)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("wantErr = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
