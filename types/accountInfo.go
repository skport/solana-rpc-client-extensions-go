package types

import (
	"fmt"
	"strconv"
)

type StakeAccount struct {
	Data struct {
		Parsed struct {
			Info struct {
				Meta  StakeAccountInfoMeta   `json:"meta"`
				Stake *StakeAccountInfoStake `json:"stake"`
			} `json:"info"`
			Type string `json:"type"`
		} `json:"parsed"`

		Program string `json:"program"`
		Space   uint64 `json:"space"`
	} `json:"data"`

	Executable bool   `json:"executable"`
	Lamports   uint64 `json:"lamports"`
	Owner      string `json:"owner"`
	RentEpoch  uint64 `json:"rentEpoch"`
}

type StakeAccountInfoMeta struct {
	Authorized struct {
		Staker     string `json:"staker"`
		Withdrawer string `json:"withdrawer"`
	} `json:"authorized"`
	Lockup struct {
		Custodian     string `json:"custodian"`
		Epoch         uint64 `json:"epoch"`
		UnixTimestamp uint64 `json:"unixTimestamp"`
	} `json:"lockup"`
	RentExemptReserve string `json:"rentExemptReserve"`
}

type StakeAccountInfoStake struct {
	CreditsObserved uint64 `json:"creditsObserved"`
	Delegation      struct {
		ActivationEpoch    string  `json:"activationEpoch"`
		DeactivationEpoch  string  `json:"deactivationEpoch"`
		Stake              string  `json:"stake"`
		Voter              string  `json:"voter"`
		WarmupCooldownRate float64 `json:"warmupCooldownRate"`
	} `json:"delegation"`
}

func (r *StakeAccount) GetInfoMeta() StakeAccountInfoMeta {
	return r.Data.Parsed.Info.Meta
}

func (r *StakeAccount) GetInfoStake() (*StakeAccountInfoStake, error) {
	// In the inactive state, r.Value.Data.Parsed.Info.Stake is nil.
	if r.Data.Parsed.Info.Stake == nil {
		return nil, fmt.Errorf("r.Value.Data.Parsed.Info.Stake is nil")
	}
	return r.Data.Parsed.Info.Stake, nil
}

func (r *StakeAccount) GetRentExemptReserve() (uint64, error) {
	m := r.GetInfoMeta()
	if m.RentExemptReserve == "" {
		return 0, fmt.Errorf("RentExemptReserve is empty")
	}
	return strconv.ParseUint(m.RentExemptReserve, 10, 64)
}

func (r *StakeAccount) GetDelegationStake() (uint64, error) {
	s, err := r.GetInfoStake()
	if err != nil {
		return 0, err
	}
	if s.Delegation.Stake == "" {
		return 0, fmt.Errorf("Delegation.Stake is empty")
	}
	return strconv.ParseUint(s.Delegation.Stake, 10, 64)
}

func (r *StakeAccount) GetActivationEpoch() (uint64, error) {
	s, err := r.GetInfoStake()
	if err != nil {
		return 0, err
	}
	if s.Delegation.ActivationEpoch == "" {
		return 0, fmt.Errorf("Delegation.ActivationEpoch is empty")
	}
	return strconv.ParseUint(s.Delegation.ActivationEpoch, 10, 64)
}

func (r *StakeAccount) GetDeactivationEpoch() (uint64, error) {
	s, err := r.GetInfoStake()
	if err != nil {
		return 0, err
	}
	if s.Delegation.DeactivationEpoch == "" {
		return 0, fmt.Errorf("Delegation.DeactivationEpoch is empty")
	}
	return strconv.ParseUint(s.Delegation.DeactivationEpoch, 10, 64)
}

type StakeHistoryAccount struct {
	Data struct {
		Parsed struct {
			Info []StakeHistoryAccountInfo `json:"info"`
			Type string                    `json:"type"`
		} `json:"parsed"`
		Program string `json:"program"`
		Space   int    `json:"space"`
	} `json:"data"`

	Executable bool   `json:"executable"`
	Lamports   uint64 `json:"lamports"`
	Owner      string `json:"owner"`
	RentEpoch  uint64 `json:"rentEpoch"`
	Space      uint64 `json:"space"`
}

type StakeHistoryAccountInfo struct {
	Epoch        int `json:"epoch"`
	StakeHistory struct {
		Activating   uint64 `json:"activating"`
		Deactivating uint64 `json:"deactivating"`
		Effective    uint64 `json:"effective"`
	} `json:"stakeHistory"`
}
