package core

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-state-types/abi"

	"github.com/filecoin-project/venus/venus-shared/actors/builtin"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/venus-cluster/venus-sector-manager/modules"
)

const (
	// TODO: The sealerAPI namespace is Venus due to historical reasons,
	// and we should consider changing it to a more appropriate name in future versions
	SealerAPINamespace = "Venus"
	MinerAPINamespace  = "Damocles.miner"
	MajorVersion       = 0
)

var Empty Meta

type Meta *struct{}

type SealerAPI interface {
	AllocateSector(context.Context, AllocateSectorSpec) (*AllocatedSector, error)

	AcquireDeals(context.Context, abi.SectorID, AcquireDealsSpec) (Deals, error)

	AssignTicket(context.Context, abi.SectorID) (Ticket, error)

	SubmitPreCommit(context.Context, AllocatedSector, PreCommitOnChainInfo, bool) (SubmitPreCommitResp, error)

	PollPreCommitState(context.Context, abi.SectorID) (PollPreCommitStateResp, error)

	// TODO：维持兼容性，应当在将来移除
	SubmitPersisted(context.Context, abi.SectorID, string) (bool, error)

	SubmitPersistedEx(ctx context.Context, sid abi.SectorID, instanceName string, isUpgrade bool) (bool, error)

	WaitSeed(context.Context, abi.SectorID) (WaitSeedResp, error)

	SubmitProof(context.Context, abi.SectorID, ProofOnChainInfo, bool) (SubmitProofResp, error)

	PollProofState(context.Context, abi.SectorID) (PollProofStateResp, error)

	ReportState(context.Context, abi.SectorID, ReportStateReq) (Meta, error)

	ReportFinalized(context.Context, abi.SectorID) (Meta, error)

	ReportAborted(context.Context, abi.SectorID, string) (Meta, error)

	// Snap
	AllocateSanpUpSector(ctx context.Context, spec AllocateSnapUpSpec) (*AllocatedSnapUpSector, error)

	SubmitSnapUpProof(ctx context.Context, sid abi.SectorID, snapupInfo SnapUpOnChainInfo) (SubmitSnapUpProofResp, error)

	// Rebuild
	AllocateRebuildSector(ctx context.Context, spec AllocateSectorSpec) (*SectorRebuildInfo, error)

	// Workers
	WorkerPing(ctx context.Context, winfo WorkerInfo) (Meta, error)

	// Store
	StoreReserveSpace(ctx context.Context, sid abi.SectorID, size uint64, candidates []string) (*StoreBasicInfo, error)

	StoreBasicInfo(ctx context.Context, instanceName string) (*StoreBasicInfo, error)

	// utils
	SealerCliAPI
}

type SealerCliAPI interface {
	ListSectors(context.Context, SectorWorkerState, SectorWorkerJob) ([]*SectorState, error)

	FindSector(ctx context.Context, state SectorWorkerState, sid abi.SectorID) (*SectorState, error)

	FindSectorsWithDeal(ctx context.Context, state SectorWorkerState, dealID abi.DealID) ([]*SectorState, error)

	ImportSector(ctx context.Context, ws SectorWorkerState, state *SectorState, override bool) (bool, error)

	RestoreSector(ctx context.Context, sid abi.SectorID, forced bool) (Meta, error)

	CheckProvable(ctx context.Context, mid abi.ActorID, postProofType abi.RegisteredPoStProof, sectors []builtin.ExtendedSectorInfo, strict, stateCheck bool) (map[abi.SectorNumber]string, error)

	SimulateWdPoSt(context.Context, address.Address, abi.RegisteredPoStProof, []builtin.ExtendedSectorInfo, abi.PoStRandomness) error

	SnapUpPreFetch(ctx context.Context, mid abi.ActorID, dlindex *uint64) (*SnapUpFetchResult, error)

	SnapUpCandidates(ctx context.Context, mid abi.ActorID) ([]*bitfield.BitField, error)

	SnapUpCancelCommitment(ctx context.Context, sid abi.SectorID) error

	ProvingSectorInfo(ctx context.Context, sid abi.SectorID) (ProvingSectorInfo, error)

	WorkerGetPingInfo(ctx context.Context, name string) (*WorkerPingInfo, error)

	WorkerPingInfoList(ctx context.Context) ([]WorkerPingInfo, error)

	SectorIndexerFind(ctx context.Context, indexType SectorIndexType, sid abi.SectorID) (SectorIndexLocation, error)

	TerminateSector(context.Context, abi.SectorID) (SubmitTerminateResp, error)

	PollTerminateSectorState(context.Context, abi.SectorID) (TerminateInfo, error)

	RemoveSector(context.Context, abi.SectorID) error

	FinalizeSector(context.Context, abi.SectorID) error

	StoreReleaseReserved(ctx context.Context, sid abi.SectorID) (bool, error)

	StoreList(ctx context.Context) ([]StoreDetailedInfo, error)

	SectorSetForRebuild(ctx context.Context, sid abi.SectorID, opt RebuildOptions) (bool, error)
}

type RandomnessAPI interface {
	GetTicket(context.Context, types.TipSetKey, abi.ChainEpoch, abi.ActorID) (Ticket, error)
	GetSeed(context.Context, types.TipSetKey, abi.ChainEpoch, abi.ActorID) (Seed, error)
	GetWindowPoStChanlleengeRand(context.Context, types.TipSetKey, abi.ChainEpoch, abi.ActorID) (WindowPoStRandomness, error)
	GetWindowPoStCommitRand(context.Context, types.TipSetKey, abi.ChainEpoch) (WindowPoStRandomness, error)
}

type MinerAPI interface {
	GetInfo(context.Context, abi.ActorID) (*MinerInfo, error)
	GetMinerConfig(context.Context, abi.ActorID) (*modules.MinerConfig, error)
}
