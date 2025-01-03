package client

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	sdkRpc "github.com/blocto/solana-go-sdk/rpc"
	"github.com/skport/solana-rpc-client-extensions-go/types"

	"golang.org/x/xerrors"
)

type GetStakeActivationResponse struct {
	Active   uint64 `json:"active"`
	Inactive uint64 `json:"inactive"`
	State    string `json:"state"`
}

const (
	// https://docs.anza.xyz/runtime/sysvars#stakehistory
	StakeHistoryAccountAddress = "SysvarStakeHistory1111111111111111111111111"
	// https://github.com/anza-xyz/solana-rpc-client-extensions/blob/aed9a86988f7f8055fe3dd3cd3e28761ad10ce04/js-v1/src/delegation.ts#L26
	stakeWarmupCooldownRate = 0.09
)

func GetStakeActivation(stakeAccountAddress string, epoch uint64, stakeAccount *types.StakeAccount, stakeHistoryAccount *types.StakeHistoryAccount) (*GetStakeActivationResponse, error) {
	var (
		effective    uint64
		activating   uint64
		deactivating uint64
	)

	// Calculates the amount of valid staking only during staking (when Info.Stake of stakeAccount is not nil).
	stakeInfo, err := stakeAccount.GetInfoStake()
	if err == nil && stakeInfo.Delegation.Stake != "" {
		effective, activating, deactivating, err = getSolanaStakeActivatingAndDeactivating(stakeAccountAddress, stakeAccount, epoch, stakeHistoryAccount)
		if err != nil {
			return nil, xerrors.Errorf("epoch: %d, stakeAccount: %s, wrap: %w", epoch, stakeAccountAddress, err)
		}
	}

	state := "inactive"
	if deactivating > 0 {
		state = "deactivating"
	} else if activating > 0 {
		state = "activating"
	} else if effective > 0 {
		state = "active"
	}

	rentExemptReserve, err := stakeAccount.GetRentExemptReserve()
	if err != nil {
		return nil, xerrors.Errorf("epoch: %d, stakeAccount: %s, wrap: %w", epoch, stakeAccountAddress, err)
	}
	inactive := stakeAccount.Lamports - effective - rentExemptReserve

	return &GetStakeActivationResponse{
		Active:   effective,
		Inactive: inactive,
		State:    state,
	}, nil
}

func getSolanaStakeActivatingAndDeactivating(stakeAccountAddress string, stakeAccount *types.StakeAccount, targetEpoch uint64, stakeHistoryAccount *types.StakeHistoryAccount) (uint64, uint64, uint64, error) {
	var (
		effective    uint64
		activating   uint64
		deactivating uint64
	)

	effective, activating, err := getSolanaStakeAndActivating(stakeAccountAddress, stakeAccount, targetEpoch, stakeHistoryAccount)
	if err != nil {
		return 0, 0, 0, xerrors.Errorf("targetEpoch: %d, wrap: %w", targetEpoch, err)
	}

	deactivationEpoch, err := stakeAccount.GetDeactivationEpoch()
	if err != nil {
		return 0, 0, 0, xerrors.Errorf("wrap: %w", err)
	}

	if targetEpoch < deactivationEpoch {
		deactivating = 0
		return effective, activating, deactivating, nil
	} else if targetEpoch == deactivationEpoch {
		activating = 0
		deactivating = effective
		return effective, activating, deactivating, nil
	}

	currentEpoch, err := stakeAccount.GetDeactivationEpoch()
	if err != nil {
		return 0, 0, 0, xerrors.Errorf("wrap: %w", err)
	}

	stakeHistoryEntry := getSolanaStakeHistoryEntry(stakeHistoryAccount, targetEpoch)

	if stakeHistoryEntry != nil {
		currentEffectiveStake := effective

		for stakeHistoryEntry != nil {
			currentEpoch++

			if stakeHistoryEntry.StakeHistory.Deactivating == 0 {
				break
			}

			// calculate weight
			currentEffectiveStakeBig := new(big.Float).SetUint64(currentEffectiveStake)
			historyDeactivatingBig := new(big.Float).SetUint64(stakeHistoryEntry.StakeHistory.Deactivating)
			weightBig := new(big.Float).Quo(currentEffectiveStakeBig, historyDeactivatingBig)

			// calculate newly not effective cluster stake
			effectiveBig := new(big.Float).SetUint64(stakeHistoryEntry.StakeHistory.Effective)
			StakeWarmupCooldownRateBig := new(big.Float).SetFloat64(stakeWarmupCooldownRate)
			newlyNotEffectiveClusterStakeBig := new(big.Float).Mul(effectiveBig, StakeWarmupCooldownRateBig)

			// calculate newly not effective stake
			r := new(big.Float).Mul(weightBig, newlyNotEffectiveClusterStakeBig)
			roundedResult, accuracy := r.Float64()
			if accuracy != big.Exact {
				return 0, 0, 0, fmt.Errorf("failed to calculate newly not effective cluster stake, stakeAccount: %s, currentEpoch: %d, roundedResult: %f, accuracy: %s",
					stakeAccountAddress,
					currentEpoch,
					roundedResult,
					accuracy.String(),
				)
			}
			newlyNotEffectiveStake := uint64(math.Max(1, math.Round(roundedResult)))

			currentEffectiveStake -= newlyNotEffectiveStake
			if currentEffectiveStake <= 0 {
				currentEffectiveStake = 0
				break
			}

			if currentEpoch >= targetEpoch {
				break
			}

			stakeHistoryEntry = getSolanaStakeHistoryEntry(stakeHistoryAccount, currentEpoch)
		}

		effective = currentEffectiveStake
		deactivating = currentEffectiveStake
		activating = 0
	} else {
		effective = 0
		activating = 0
		deactivating = 0
	}

	return effective, activating, deactivating, nil
}

func getSolanaStakeAndActivating(stakeAccountAddress string, stakeAccount *types.StakeAccount, targetEpoch uint64, stakeHistoryAccount *types.StakeHistoryAccount) (uint64, uint64, error) {
	var (
		effective  uint64
		activating uint64
	)

	activationEpoch, err := stakeAccount.GetActivationEpoch()
	if err != nil {
		return 0, 0, xerrors.Errorf("wrap: %w", err)
	}
	deactivationEpoch, err := stakeAccount.GetDeactivationEpoch()
	if err != nil {
		return 0, 0, xerrors.Errorf("wrap: %w", err)
	}
	delegationStake, err := stakeAccount.GetDelegationStake()
	if err != nil {
		return 0, 0, xerrors.Errorf("wrap: %w", err)
	}

	if activationEpoch == deactivationEpoch {
		return 0, 0, nil
	} else if targetEpoch == activationEpoch {
		return 0, delegationStake, nil
	} else if targetEpoch < activationEpoch {
		return 0, 0, nil
	}

	currentEpoch := activationEpoch
	stakeHistoryEntry := getSolanaStakeHistoryEntry(stakeHistoryAccount, targetEpoch)

	if stakeHistoryEntry != nil {
		currentEffectiveStake := uint64(0)

		for stakeHistoryEntry != nil {
			currentEpoch++

			remaining := delegationStake - currentEffectiveStake

			// calculate weight
			remainingBig := new(big.Float).SetUint64(remaining)
			activatingBig := new(big.Float).SetUint64(activating)
			weightBig := new(big.Float).Quo(remainingBig, activatingBig)

			// calculate newly effective cluster stake
			effectiveBig := new(big.Float).SetUint64(stakeHistoryEntry.StakeHistory.Effective)
			StakeWarmupCooldownRateBig := new(big.Float).SetFloat64(stakeWarmupCooldownRate)
			newlyEffectiveClusterStakeBig := new(big.Float).Mul(effectiveBig, StakeWarmupCooldownRateBig)

			// calculate newly effective stake
			r := new(big.Float).Mul(weightBig, newlyEffectiveClusterStakeBig)
			roundedResult, accuracy := r.Float64()
			if accuracy != big.Exact {
				return 0, 0, fmt.Errorf("failed to calculate newly effective cluster stake, stakeAccount: %s, currentEpoch: %d, roundedResult: %f, accuracy: %s",
					stakeAccountAddress,
					currentEpoch,
					roundedResult,
					accuracy.String(),
				)
			}
			newlyEffectiveStake := uint64(math.Max(1, math.Round(roundedResult)))

			currentEffectiveStake += newlyEffectiveStake
			if currentEffectiveStake >= delegationStake {
				currentEffectiveStake = delegationStake
				break
			}

			if currentEpoch >= targetEpoch || currentEpoch >= deactivationEpoch {
				break
			}

			stakeHistoryEntry = getSolanaStakeHistoryEntry(stakeHistoryAccount, currentEpoch)
		}

		effective = currentEffectiveStake
		activating = delegationStake - currentEffectiveStake
	} else {
		effective = delegationStake
		activating = 0
	}

	return effective, activating, nil
}

func getSolanaStakeHistoryEntry(r *types.StakeHistoryAccount, targetEpoch uint64) *types.StakeHistoryAccountInfo {
	for _, entry := range r.Data.Parsed.Info {
		if uint64(entry.Epoch) == targetEpoch {
			return &entry
		}
	}
	return nil
}

func ConvertStakeAccountInfo(stakeAccountInfo sdkRpc.JsonRpcResponse[sdkRpc.ValueWithContext[sdkRpc.AccountInfo]]) (*types.StakeAccount, error) {
	b, err := json.Marshal(stakeAccountInfo.Result.Value)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal: %w", err)
	}

	var stakeAccount *types.StakeAccount
	if err := json.Unmarshal(b, &stakeAccount); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal to StakeAccountInfoResponse: %v", err)
	}

	return stakeAccount, nil
}

func ConvertStakeHistoryAccountInfo(stakeHistoryAccountInfo sdkRpc.JsonRpcResponse[sdkRpc.ValueWithContext[sdkRpc.AccountInfo]]) (*types.StakeHistoryAccount, error) {
	b, err := json.Marshal(stakeHistoryAccountInfo.Result.Value)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal: %w", err)
	}

	var stakeHistoryAccount *types.StakeHistoryAccount
	if err := json.Unmarshal(b, &stakeHistoryAccount); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal to StakeHistoryAccountInfoResponse: %v", err)
	}

	if stakeHistoryAccount.Data.Parsed.Info == nil {
		return nil, xerrors.Errorf("stakeHistoryAccount.Data.Parsed.Info is nil")
	}

	return stakeHistoryAccount, nil
}
