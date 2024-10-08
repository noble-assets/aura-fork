package aura

import (
	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ondoprotocol/usdy-noble/v2/keeper"
	"github.com/ondoprotocol/usdy-noble/v2/types"
	"github.com/ondoprotocol/usdy-noble/v2/types/blocklist"
)

func InitGenesis(ctx sdk.Context, k *keeper.Keeper, addressCodec address.Codec, genesis types.GenesisState) {
	k.SetBlocklistOwner(ctx, genesis.BlocklistState.Owner)
	k.SetBlocklistPendingOwner(ctx, genesis.BlocklistState.PendingOwner)
	for _, account := range genesis.BlocklistState.BlockedAddresses {
		address, _ := addressCodec.StringToBytes(account)
		k.SetBlockedAddress(ctx, address)
	}

	k.SetPaused(ctx, genesis.Paused)
	k.SetOwner(ctx, genesis.Owner)
	k.SetPendingOwner(ctx, genesis.PendingOwner)
	for _, burner := range genesis.Burners {
		k.SetBurner(ctx, burner.Address, burner.Allowance)
	}
	for _, minter := range genesis.Minters {
		k.SetMinter(ctx, minter.Address, minter.Allowance)
	}
	for _, pauser := range genesis.Pausers {
		k.SetPauser(ctx, pauser)
	}
}

func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		BlocklistState: blocklist.GenesisState{
			Owner:            k.GetBlocklistOwner(ctx),
			PendingOwner:     k.GetBlocklistPendingOwner(ctx),
			BlockedAddresses: k.GetBlockedAddresses(ctx),
		},
		Paused:       k.GetPaused(ctx),
		Owner:        k.GetOwner(ctx),
		PendingOwner: k.GetPendingOwner(ctx),
		Burners:      k.GetBurners(ctx),
		Minters:      k.GetMinters(ctx),
		Pausers:      k.GetPausers(ctx),
	}
}
