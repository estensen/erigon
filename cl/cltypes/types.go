package cltypes

import (
	"bytes"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/cl/utils"
	"github.com/ledgerwatch/erigon/core/types"
)

/*
 * Block body for Consensus Layer, we only care about its hash and execution payload.
 */
type BeaconBodyBellatrix struct {
	RandaoReveal      [96]byte `ssz-size:"96"`
	Eth1Data          *Eth1Data
	Graffiti          []byte                 `ssz-size:"32"`
	ProposerSlashings []*ProposerSlashing    `ssz-max:"16"`
	AttesterSlashings []*AttesterSlashing    `ssz-max:"2"`
	Attestations      []*Attestation         `ssz-max:"128"`
	Deposits          []*Deposit             `ssz-max:"16"`
	VoluntaryExits    []*SignedVoluntaryExit `ssz-max:"16"`
	SyncAggregate     *SyncAggregate
	ExecutionPayload  *Eth1Block
}

/*
 * Sync committe public keys and their aggregate public keys, we use array of pubKeys.
 */
type SyncCommittee struct {
	PubKeys            [][48]byte `ssz-size:"512,48"`
	AggregatePublicKey [48]byte   `ssz-size:"48"`
}

func (s *SyncCommittee) Equal(s2 *SyncCommittee) bool {
	if !bytes.Equal(s.AggregatePublicKey[:], s2.AggregatePublicKey[:]) {
		return false
	}
	if len(s.PubKeys) != len(s2.PubKeys) {
		return false
	}
	for i := range s.PubKeys {
		if !bytes.Equal(s.PubKeys[i][:], s2.PubKeys[i][:]) {
			return false
		}
	}
	return true
}

// LightClientBootstrap is used to bootstrap the lightclient from checkpoint sync.
type LightClientBootstrap struct {
	Header                     *BeaconBlockHeader
	CurrentSyncCommittee       *SyncCommittee
	CurrentSyncCommitteeBranch []libcommon.Hash
}

func (l *LightClientBootstrap) UnmarshalSSZWithVersion(buf []byte, _ int) error {
	return l.UnmarshalSSZ(buf)
}

// LightClientUpdate is used to update the sync committee every 27 hours.
type LightClientUpdate struct {
	AttestedHeader          *BeaconBlockHeader
	NextSyncCommitee        *SyncCommittee
	NextSyncCommitteeBranch []libcommon.Hash
	FinalizedHeader         *BeaconBlockHeader
	FinalityBranch          []libcommon.Hash
	SyncAggregate           *SyncAggregate
	SignatureSlot           uint64
}

func (l *LightClientUpdate) UnmarshalSSZWithVersion(buf []byte, _ int) error {
	return l.UnmarshalSSZ(buf)
}

func (l *LightClientUpdate) HasNextSyncCommittee() bool {
	return l.NextSyncCommitee != nil
}

func (l *LightClientUpdate) IsFinalityUpdate() bool {
	return l.FinalityBranch != nil
}

func (l *LightClientUpdate) HasSyncFinality() bool {
	return l.FinalizedHeader != nil &&
		utils.SlotToPeriod(l.AttestedHeader.Slot) == utils.SlotToPeriod(l.FinalizedHeader.Slot)
}

// LightClientFinalityUpdate is used to update the sync aggreggate every 6 minutes.
type LightClientFinalityUpdate struct {
	AttestedHeader  *BeaconBlockHeader
	FinalizedHeader *BeaconBlockHeader
	FinalityBranch  []libcommon.Hash `ssz-size:"6,32"`
	SyncAggregate   *SyncAggregate
	SignatureSlot   uint64
}

func (l *LightClientFinalityUpdate) UnmarshalSSZWithVersion(buf []byte, _ int) error {
	return l.UnmarshalSSZ(buf)
}

// LightClientOptimisticUpdate is used for verifying N-1 block.
type LightClientOptimisticUpdate struct {
	AttestedHeader *BeaconBlockHeader
	SyncAggregate  *SyncAggregate
	SignatureSlot  uint64
}

func (l *LightClientOptimisticUpdate) UnmarshalSSZWithVersion(buf []byte, _ int) error {
	return l.UnmarshalSSZ(buf)
}

// Fork data, contains if we were on bellatrix/alteir/phase0 and transition epoch. NOT USED.
type Fork struct {
	PreviousVersion [4]byte `ssz-size:"4" `
	CurrentVersion  [4]byte `ssz-size:"4" `
	Epoch           uint64
}

// Validator, contains if we were on bellatrix/alteir/phase0 and transition epoch.
type Validator struct {
	PublicKey                  [48]byte `ssz-size:"48"`
	WithdrawalCredentials      []byte   `ssz-size:"32"`
	EffectiveBalance           uint64
	Slashed                    bool
	ActivationEligibilityEpoch uint64
	ActivationEpoch            uint64
	ExitEpoch                  uint64
	WithdrawableEpoch          uint64
}

// BellatrixBeaconState is the bellatrix beacon state.
type BeaconStateBellatrix struct {
	GenesisTime                  uint64
	GenesisValidatorsRoot        [32]byte `ssz-size:"32"`
	Slot                         uint64
	Fork                         *Fork
	LatestBlockHeader            *BeaconBlockHeader
	BlockRoots                   [][32]byte `ssz-size:"8192,32"`
	StateRoots                   [][32]byte `ssz-size:"8192,32"`
	HistoricalRoots              [][32]byte `ssz-max:"16777216" ssz-size:"?,32"`
	Eth1Data                     *Eth1Data
	Eth1DataVotes                []*Eth1Data `ssz-max:"2048"`
	Eth1DepositIndex             uint64
	Validators                   []*Validator `ssz-max:"1099511627776"`
	Balances                     []uint64     `ssz-max:"1099511627776"`
	RandaoMixes                  [][32]byte   `ssz-size:"65536,32"`
	Slashings                    []uint64     `ssz-size:"8192"`
	PreviousEpochParticipation   []byte       `ssz-max:"1099511627776"`
	CurrentEpochParticipation    []byte       `ssz-max:"1099511627776"`
	JustificationBits            []byte       `ssz-size:"1"`
	PreviousJustifiedCheckpoint  *Checkpoint
	CurrentJustifiedCheckpoint   *Checkpoint
	FinalizedCheckpoint          *Checkpoint
	InactivityScores             []uint64 `ssz-max:"1099511627776"`
	CurrentSyncCommittee         *SyncCommittee
	NextSyncCommittee            *SyncCommittee
	LatestExecutionPayloadHeader *types.Header
}

// BlockRoot retrieves a the state block root from the state.
func (b *BeaconStateBellatrix) BlockRoot() ([32]byte, error) {
	stateRoot, err := b.HashTreeRoot()
	if err != nil {
		return [32]byte{}, nil
	}
	// We make a temporary header for block root computation
	tempHeader := &BeaconBlockHeader{
		Slot:          b.LatestBlockHeader.Slot,
		ProposerIndex: b.LatestBlockHeader.ProposerIndex,
		ParentRoot:    b.LatestBlockHeader.ParentRoot,
		BodyRoot:      b.LatestBlockHeader.BodyRoot,
		Root:          stateRoot,
	}
	return tempHeader.HashTreeRoot()
}
