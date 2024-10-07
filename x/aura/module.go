package aura

import (
	"context"
	"encoding/json"
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	blocklistv1 "github.com/ondoprotocol/usdy-noble/v2/api/blocklist/v1"
	modulev1 "github.com/ondoprotocol/usdy-noble/v2/api/module/v1"
	aurav1 "github.com/ondoprotocol/usdy-noble/v2/api/v1"
	"github.com/ondoprotocol/usdy-noble/v2/x/aura/keeper"
	"github.com/ondoprotocol/usdy-noble/v2/x/aura/types"
	"github.com/ondoprotocol/usdy-noble/v2/x/aura/types/blocklist"
)

// ConsensusVersion defines the current x/aura module consensus version.
const ConsensusVersion = 1

var (
	_ module.AppModuleBasic      = AppModule{}
	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasGenesisBasics    = AppModuleBasic{}
	_ module.HasServices         = AppModule{}
)

//

type AppModuleBasic struct {
	addressCodec address.Codec
}

func NewAppModuleBasic(addressCodec address.Codec) AppModuleBasic {
	return AppModuleBasic{addressCodec: addressCodec}
}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}

	if err := blocklist.RegisterQueryHandlerClient(context.Background(), mux, blocklist.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesis); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesis.Validate(b.addressCodec)
}

//

type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

func NewAppModule(addressCodec address.Codec, keeper *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(addressCodec),
		keeper:         keeper,
	}
}

func (AppModule) IsOnePerModuleType() {}

func (AppModule) IsAppModule() {}

func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

func (m AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesis)

	InitGenesis(ctx, m.keeper, m.addressCodec, genesis)
}

func (m AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := ExportGenesis(ctx, m.keeper)
	return cdc.MustMarshalJSON(genesis)
}

func (m AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(m.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(m.keeper))

	blocklist.RegisterMsgServer(cfg.MsgServer(), keeper.NewBlocklistMsgServer(m.keeper))
	blocklist.RegisterQueryServer(cfg.QueryServer(), keeper.NewBlocklistQueryServer(m.keeper))
}

//

func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: aurav1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Burn",
					Use:       "burn [from] [amount]",
					Short:     "Transaction that burns tokens",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "from"},
						{ProtoField: "amount"},
					},
				},
				{
					RpcMethod: "Mint",
					Use:       "mint [to] [amount]",
					Short:     "Transaction that mints tokens",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "to"},
						{ProtoField: "amount"},
					},
				},
				{
					RpcMethod: "Pause",
					Use:       "pause",
					Short:     "Transaction that pauses the module",
				},
				{
					RpcMethod: "Unpause",
					Use:       "unpause",
					Short:     "Transaction that unpauses the module",
				},
				{
					RpcMethod:      "TransferOwnership",
					Use:            "transfer-ownership [new-owner]",
					Short:          "Transfer ownership of module",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "new_owner"}},
				},
				{
					RpcMethod: "AcceptOwnership",
					Use:       "accept-ownership",
					Short:     "Accept ownership of module",
					Long:      "Accept ownership of module, assuming there is an pending ownership transfer",
				},
				{
					RpcMethod: "AddBurner",
					Use:       "add-burner [burner] [allowance]",
					Short:     "Add a new burner with an initial allowance",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "burner"},
						{ProtoField: "allowance"},
					},
				},
				{
					RpcMethod:      "RemoveBurner",
					Use:            "remove-burner [burner]",
					Short:          "Remove an existing burner",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "burner"}},
				},
				{
					RpcMethod: "SetBurnerAllowance",
					Use:       "set-burner-allowance [burner] [allowance]",
					Short:     "Set an existing burner's allowance",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "burner"},
						{ProtoField: "allowance"},
					},
				},
				{
					RpcMethod: "AddMinter",
					Use:       "add-minter [minter] [allowance]",
					Short:     "Add a new minter with an initial allowance",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "minter"},
						{ProtoField: "allowance"},
					},
				},
				{
					RpcMethod:      "RemoveMinter",
					Use:            "remove-minter [minter]",
					Short:          "Remove an existing minter",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "minter"}},
				},
				{
					RpcMethod: "SetMinterAllowance",
					Use:       "set-minter-allowance [minter] [allowance]",
					Short:     "Set an existing minter's allowance",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "minter"},
						{ProtoField: "allowance"},
					},
				},
				{
					RpcMethod:      "AddPauser",
					Use:            "add-pauser [pauser]",
					Short:          "Add a new pauser",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pauser"}},
				},
				{
					RpcMethod:      "RemovePauser",
					Use:            "remove-pauser [pauser]",
					Short:          "Remove an existing pauser",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pauser"}},
				},
				{
					RpcMethod:      "AddBlockedChannel",
					Use:            "add-blocked-channel [channel]",
					Short:          "Add a new blocked channel",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "channel"}},
				},
				{
					RpcMethod:      "RemoveBlockedChannel",
					Use:            "remove-blocked-channel [channel]",
					Short:          "Remove an existing blocked channel",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "channel"}},
				},
			},
			SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
				"blocklist": {
					Service: blocklistv1.Msg_ServiceDesc.ServiceName,
					Short:   "Transactions commands for the blocklist submodule",
					RpcCommandOptions: []*autocliv1.RpcCommandOptions{
						{
							RpcMethod:      "TransferOwnership",
							Use:            "transfer-ownership [new-owner]",
							Short:          "Transfer ownership of submodule",
							PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "new_owner"}},
						},
						{
							RpcMethod: "AcceptOwnership",
							Use:       "accept-ownership",
							Short:     "Accept ownership of submodule",
							Long:      "Accept ownership of submodule, assuming there is an pending ownership transfer",
						},
						{
							RpcMethod: "AddToBlocklist",
							Use:       "add-to-blocklist [addresses ...]",
							Short:     "Add a list of accounts to the blocklist",
							PositionalArgs: []*autocliv1.PositionalArgDescriptor{
								{ProtoField: "accounts", Varargs: true},
							},
						},
						{
							RpcMethod: "RemoveFromBlocklist",
							Use:       "remove-from-blocklist [addresses ...]",
							Short:     "Remove a list of accounts from the blocklist",
							PositionalArgs: []*autocliv1.PositionalArgDescriptor{
								{ProtoField: "accounts", Varargs: true},
							},
						},
					},
				},
			},
		},
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: aurav1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Denom",
					Use:       "denom",
					Short:     "Query the module's denom",
				},
				{
					RpcMethod: "Paused",
					Use:       "paused",
					Short:     "Query if the module is paused",
				},
				{
					RpcMethod: "Owner",
					Use:       "owner",
					Short:     "Query the module's owner",
				},
				{
					RpcMethod: "Burners",
					Use:       "burners",
					Short:     "Query the module's burners",
				},
				{
					RpcMethod: "Minters",
					Use:       "minters",
					Short:     "Query the module's minters",
				},
				{
					RpcMethod: "Pausers",
					Use:       "pausers",
					Short:     "Query the module's pausers",
				},
				{
					RpcMethod: "BlockedChannels",
					Use:       "blocked-channels",
					Short:     "Query the blocked channels",
				},
			},
			SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
				"blocklist": {
					Service: blocklistv1.Query_ServiceDesc.ServiceName,
					Short:   "Querying commands for the blocklist submodule",
					RpcCommandOptions: []*autocliv1.RpcCommandOptions{
						{
							RpcMethod: "Owner",
							Use:       "owner",
							Short:     "Query the submodule's owner",
						},
						{
							RpcMethod: "Addresses",
							Use:       "addresses",
							Short:     "Query for all blocked addresses",
						},
						{
							RpcMethod:      "Address",
							Use:            "address [address]",
							Short:          "Query if an address is blocked",
							PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
						},
					},
				},
			},
		},
	}
}

//

func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	StoreService store.KVStoreService
	EventService event.Service

	Cdc          codec.Codec
	AddressCodec address.Codec
	BankKeeper   types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper      *keeper.Keeper
	Module      appmodule.AppModule
	Restriction banktypes.SendRestrictionFn
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(
		in.Config.Denom,
		in.StoreService,
		in.EventService,
		in.AddressCodec,
		in.BankKeeper,
	)
	m := NewAppModule(in.AddressCodec, k)

	return ModuleOutputs{Keeper: k, Module: m, Restriction: k.SendRestrictionFn}
}
