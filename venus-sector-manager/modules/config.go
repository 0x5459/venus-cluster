package modules

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/ipfs-force-community/venus-cluster/venus-sector-manager/pkg/messager"

	"github.com/ipfs-force-community/venus-cluster/venus-sector-manager/pkg/confmgr"
	"github.com/ipfs-force-community/venus-cluster/venus-sector-manager/pkg/objstore"
)

func init() {
	fake, err := address.NewFromString("f1abjxfbp274xpdqcpuaykwkfb43omjotacm2p3za")
	if err != nil {
		panic(fmt.Errorf("parse fake address: %w", err))
	}

	fakeAddress = MustAddress(fake)
}

type SafeConfig struct {
	*Config
	sync.Locker
}

func (sc *SafeConfig) MustCommonConfig() CommonConfig {
	sc.Lock()
	defer sc.Unlock()

	return sc.Common
}

func (sc *SafeConfig) MustMinerConfig(mid abi.ActorID) MinerConfig {
	mc, err := sc.MinerConfig(mid)
	if err != nil {
		panic(err)
	}

	return mc
}

func (sc *SafeConfig) MinerConfig(mid abi.ActorID) (MinerConfig, error) {
	sc.Lock()
	defer sc.Unlock()

	for i := range sc.Miners {
		if sc.Miners[i].Actor == mid {
			return sc.Miners[i], nil
		}
	}

	return MinerConfig{}, fmt.Errorf("config for miner actor %d not found", mid)
}

var fakeAddress MustAddress

func GetFakeAddress() MustAddress {
	return fakeAddress
}

const ConfigKey = "sector-manager"

type CommonAPIConfig struct {
	Chain              string
	Messager           string
	Market             string
	Gateway            []string
	Token              string
	ChainEventInterval Duration
}

func defaultCommonAPIConfig(example bool) CommonAPIConfig {
	cfg := CommonAPIConfig{
		ChainEventInterval: Duration(time.Minute),
	}
	if example {
		cfg.Chain = "/ip4/{api_host}/tcp/{api_port}"
		cfg.Messager = "/ip4/{api_host}/tcp/{api_port}"
		cfg.Market = "/ip4/{api_host}/tcp/{api_port}"
		cfg.Gateway = []string{"/ip4/{api_host}/tcp/{api_port}"}
		cfg.Token = "{some token}"
	}
	return cfg
}

type PluginConfig struct {
	Dir string
}

func DefaultPluginConfig() *PluginConfig {
	return &PluginConfig{
		Dir: "",
	}
}

type DBConfig struct {
	Driver string
	Badger *KVStoreBadgerDBConfig
	Mongo  *KVStoreMongoDBConfig
	Plugin *KVStorePluginDBConfig
}

func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		Driver: "badger",
		Badger: defaultBadgerDBConfig(),
		Mongo:  nil,
		Plugin: nil,
	}
}

type KVStoreBadgerDBConfig struct {
	BaseDir string
}

func defaultBadgerDBConfig() *KVStoreBadgerDBConfig {
	return &KVStoreBadgerDBConfig{
		BaseDir: "",
	}
}

type KVStoreMongoDBConfig struct {
	Enable       bool // For compatibility with v0.5
	DSN          string
	DatabaseName string
}

type KVStorePluginDBConfig struct {
	PluginName string
	Meta       map[string]string
}

type PieceStoreConfig struct {
	Name       string
	Path       string
	Meta       map[string]string
	Plugin     string // For compatibility with v0.5
	PluginName string
}

type PersistStoreConfig struct {
	objstore.Config
	objstore.StoreSelectPolicy

	Plugin     string // For compatibility with v0.5
	PluginName string
}

type ProvingConfig struct {
	// Maximum number of sector checks to run in parallel. (0 = unlimited)
	//
	// WARNING: Setting this value too high may make the node crash by running out of stack
	// WARNING: Setting this value too low may make sector challenge reading much slower, resulting in failed PoSt due
	// to late submission.
	ParallelCheckLimit int

	// Maximum amount of time a proving pre-check can take for a sector. If the check times out the sector will be skipped
	//
	// WARNING: Setting this value too low risks in sectors being skipped even though they are accessible, just reading the
	// test challenge took longer than this timeout
	// WARNING: Setting this value too high risks missing PoSt deadline in case IO operations related to this sector are
	// blocked (e.g. in case of disconnected NFS mount)
	SingleCheckTimeout Duration

	// Maximum amount of time a proving pre-check can take for an entire partition. If the check times out, sectors in
	// the partition which didn't get checked on time will be skipped
	//
	// WARNING: Setting this value too low risks in sectors being skipped even though they are accessible, just reading the
	// test challenge took longer than this timeout
	// WARNING: Setting this value too high risks missing PoSt deadline in case IO operations related to this partition are
	// blocked or slow
	PartitionCheckTimeout Duration
}

func defaultProvingConfig() ProvingConfig {
	cfg := ProvingConfig{
		ParallelCheckLimit:    128,
		PartitionCheckTimeout: Duration(20 * time.Minute),
		SingleCheckTimeout:    Duration(10 * time.Minute),
	}
	return cfg
}

type CommonConfig struct {
	API           CommonAPIConfig
	Plugins       *PluginConfig
	PieceStores   []PieceStoreConfig
	PersistStores []PersistStoreConfig
	MongoKVStore  *KVStoreMongoDBConfig // For compatibility with v0.5
	DB            *DBConfig
	Proving       ProvingConfig
}

func exampleFilestoreConfig() objstore.Config {
	cfg := objstore.DefaultConfig("{store_path}", false)
	cfg.Name = "{store_name}"
	cfg.Meta["SomeKey"] = "SomeValue"
	return cfg
}

func defaultCommonConfig(example bool) CommonConfig {
	cfg := CommonConfig{
		API:           defaultCommonAPIConfig(example),
		PieceStores:   []PieceStoreConfig{},
		PersistStores: []PersistStoreConfig{},
		MongoKVStore:  nil,
		DB:            DefaultDBConfig(),
		Proving:       defaultProvingConfig(),
	}

	if example {
		exampleCfg := exampleFilestoreConfig()
		cfg.PieceStores = append(cfg.PieceStores, PieceStoreConfig{
			Name:       exampleCfg.Name,
			Path:       exampleCfg.Path,
			Meta:       exampleCfg.Meta,
			PluginName: "s3store",
		})

		cfg.PersistStores = append(cfg.PersistStores, PersistStoreConfig{
			Config: objstore.Config{
				Name: exampleCfg.Name,
				Path: exampleCfg.Path,
				Meta: exampleCfg.Meta,
			},
			StoreSelectPolicy: objstore.StoreSelectPolicy{AllowMiners: []abi.ActorID{1, 2}, DenyMiners: []abi.ActorID{3, 4}},
			PluginName:        "s3store",
		})
	}

	return cfg
}

type FeeConfig struct {
	//included in msg send spec
	GasOverEstimation float64
	GasOverPremium    float64
	//set to msg directly
	GasFeeCap FIL
	MaxFeeCap FIL //兼容老字段， FeeConfig用在embed字段， 使用TextUnmarshaler会影响上层结构体解析
}

func (feeCfg *FeeConfig) GetGasFeeCap() FIL {
	gasFeeCap := feeCfg.GasFeeCap
	if (*big.Int)(&feeCfg.GasFeeCap).NilOrZero() && !(*big.Int)(&feeCfg.MaxFeeCap).NilOrZero() {
		//if not set GasFeeCap but set MaxFeecap use MaxFeecap as GasFeecap
		gasFeeCap = feeCfg.MaxFeeCap
	}
	return gasFeeCap
}

func (feeCfg *FeeConfig) GetSendSpec() messager.MsgMeta {
	return messager.MsgMeta{
		GasOverEstimation: feeCfg.GasOverEstimation,
		GasOverPremium:    feeCfg.GasOverPremium,
	}
}

func defaultFeeConfig() FeeConfig {
	return FeeConfig{
		GasOverEstimation: 1.2,
		GasOverPremium:    0,
		GasFeeCap:         NanoFIL.Mul(5),
	}
}

type MinerSectorConfig struct {
	InitNumber   uint64
	MinNumber    *uint64
	MaxNumber    *uint64
	Enabled      bool
	EnableDeals  bool
	LifetimeDays uint64
	Verbose      bool
}

func defaultMinerSectorConfig(example bool) MinerSectorConfig {
	cfg := MinerSectorConfig{
		InitNumber:   0,
		Enabled:      true,
		LifetimeDays: 540,
		Verbose:      false,
	}

	if example {
		min := uint64(10)
		max := uint64(1_000_000)
		cfg.MinNumber = &min
		cfg.MaxNumber = &max
	}
	return cfg
}

type MinerSnapUpRetryConfig struct {
	MaxAttempts      *int
	PollInterval     Duration
	APIFailureWait   Duration
	LocalFailureWait Duration
}

func defaultMinerSnapUpRetryConfig(example bool) MinerSnapUpRetryConfig {
	cfg := MinerSnapUpRetryConfig{
		MaxAttempts:      nil,
		PollInterval:     Duration(3 * time.Minute),
		APIFailureWait:   Duration(3 * time.Minute),
		LocalFailureWait: Duration(3 * time.Minute),
	}

	if example {
		maxAttempts := 10
		cfg.MaxAttempts = &maxAttempts
	}

	return cfg
}

type MinerSnapUpConfig struct {
	Enabled  bool
	Sender   MustAddress
	SendFund bool
	FeeConfig

	MessageConfidential *abi.ChainEpoch
	ReleaseCondidential *abi.ChainEpoch

	MessageConfidence abi.ChainEpoch
	ReleaseConfidence abi.ChainEpoch

	Retry MinerSnapUpRetryConfig
}

func (m *MinerSnapUpConfig) GetMessageConfidence() abi.ChainEpoch {
	if m.MessageConfidential != nil {
		return *m.MessageConfidential
	}

	return m.MessageConfidence
}

func (m *MinerSnapUpConfig) GetReleaseConfidence() abi.ChainEpoch {
	if m.ReleaseCondidential != nil {
		return *m.ReleaseCondidential
	}

	return m.ReleaseConfidence
}

func defaultMinerSnapUpConfig(example bool) MinerSnapUpConfig {
	cfg := MinerSnapUpConfig{
		Enabled:           false,
		SendFund:          true,
		FeeConfig:         defaultFeeConfig(),
		MessageConfidence: 15,
		ReleaseConfidence: 30,
		Retry:             defaultMinerSnapUpRetryConfig(example),
	}

	if example {
		cfg.Sender = fakeAddress
	}

	return cfg
}

type MinerCommitmentConfig struct {
	Confidence int64
	Pre        MinerCommitmentPolicyConfig
	Prove      MinerCommitmentPolicyConfig
	Terminate  MinerCommitmentPolicyConfig
}

func defaultMinerCommitmentConfig(example bool) MinerCommitmentConfig {
	cfg := MinerCommitmentConfig{
		Confidence: 10,
		Pre:        defaultMinerCommitmentPolicyConfig(example),
		Prove:      defaultMinerCommitmentPolicyConfig(example),
		Terminate:  defaultMinerCommitmentPolicyConfig(example),
	}

	cfg.Terminate.Batch.Threshold = 5

	return cfg
}

type MinerCommitmentPolicyConfig struct {
	Sender   MustAddress
	SendFund bool
	FeeConfig
	Batch MinerCommitmentBatchPolicyConfig
}

func defaultMinerCommitmentPolicyConfig(example bool) MinerCommitmentPolicyConfig {
	cfg := MinerCommitmentPolicyConfig{
		SendFund:  true,
		FeeConfig: defaultFeeConfig(),
		Batch:     defaultMinerCommitmentBatchPolicyConfig(),
	}

	if example {
		cfg.Sender = fakeAddress
	}

	return cfg
}

type MinerCommitmentBatchPolicyConfig struct {
	Enabled       bool
	Threshold     int
	MaxWait       Duration
	CheckInterval Duration
	FeeConfig
}

func defaultMinerCommitmentBatchPolicyConfig() MinerCommitmentBatchPolicyConfig {
	cfg := MinerCommitmentBatchPolicyConfig{
		Enabled:       false,
		Threshold:     16,
		MaxWait:       Duration(time.Hour),
		CheckInterval: Duration(time.Minute),
		FeeConfig:     defaultFeeConfig(),
	}

	return cfg
}

type MinerPoStConfig struct {
	Sender      MustAddress
	Enabled     bool
	StrictCheck bool
	Parallel    bool
	FeeConfig
	Confidence                      uint64
	SubmitConfidence                uint64
	ChallengeConfidence             uint64
	MaxRecoverSectorLimit           uint64
	MaxPartitionsPerPoStMessage     uint64
	MaxPartitionsPerRecoveryMessage uint64
}

func DefaultMinerPoStConfig(example bool) MinerPoStConfig {
	cfg := MinerPoStConfig{
		Enabled:                         true,
		StrictCheck:                     true,
		Parallel:                        false,
		FeeConfig:                       defaultFeeConfig(),
		Confidence:                      10,
		SubmitConfidence:                0,
		ChallengeConfidence:             0,
		MaxRecoverSectorLimit:           0,
		MaxPartitionsPerPoStMessage:     0,
		MaxPartitionsPerRecoveryMessage: 0,
	}

	if example {
		cfg.Sender = fakeAddress
	}

	return cfg
}

type MinerProofConfig struct {
	Enabled bool
}

func defaultMinerProofConfig() MinerProofConfig {
	return MinerProofConfig{
		Enabled: false,
	}
}

type MinerSealingConfig struct {
	SealingEpochDuration int64
}

func defaultMinerSealingConfig() MinerSealingConfig {
	return MinerSealingConfig{
		SealingEpochDuration: 0,
	}
}

type MinerConfig struct {
	Actor      abi.ActorID
	Sector     MinerSectorConfig
	SnapUp     MinerSnapUpConfig
	Commitment MinerCommitmentConfig
	PoSt       MinerPoStConfig
	Proof      MinerProofConfig
	Sealing    MinerSealingConfig
}

func DefaultMinerConfig(example bool) MinerConfig {
	cfg := MinerConfig{
		Sector:     defaultMinerSectorConfig(example),
		SnapUp:     defaultMinerSnapUpConfig(example),
		Commitment: defaultMinerCommitmentConfig(example),
		PoSt:       DefaultMinerPoStConfig(example),
		Proof:      defaultMinerProofConfig(),
		Sealing:    defaultMinerSealingConfig(),
	}

	if example {
		cfg.Actor = 10086
	}

	return cfg
}

var _ confmgr.ConfigUnmarshaller = (*Config)(nil)

type Config struct {
	Common CommonConfig
	Miners []MinerConfig
}

func (c *Config) UnmarshalConfig(data []byte) error {
	primitive := struct {
		Common CommonConfig
		Miners []toml.Primitive
	}{
		Common: defaultCommonConfig(false),
	}

	meta, err := toml.NewDecoder(bytes.NewReader(data)).Decode(&primitive)
	if err != nil {
		return fmt.Errorf("toml.Unmarshal to primitive: %w", err)
	}

	miners := make([]MinerConfig, 0, len(primitive.Miners))
	for i, pm := range primitive.Miners {
		mcfg := DefaultMinerConfig(false)
		err := meta.PrimitiveDecode(pm, &mcfg)
		if err != nil {
			return fmt.Errorf("decode primitive to miner config #%d: %w", i, err)
		}

		miners = append(miners, mcfg)
	}

	c.Common = primitive.Common
	c.Miners = miners
	return nil
}

func DefaultConfig(example bool) Config {
	cfg := Config{
		Common: defaultCommonConfig(example),
	}

	if example {
		cfg.Miners = append(cfg.Miners, DefaultMinerConfig(example))
	}

	return cfg
}
