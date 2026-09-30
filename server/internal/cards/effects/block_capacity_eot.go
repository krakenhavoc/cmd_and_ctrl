package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_capacity_eot.go is the card half of #1715, the follow-ups to
// #1706's multi-blocking (ADR 0045's amendment of 2026-09-28,
// Decisions 64-67). multi_block.go has the STATIC builders; this file
// has the RESOLVING-effect ones and the two small readers the new
// cards share:
//
//	BlockCapacityUntilEOT{Target: t, Additional: 1}.Apply(ctx) // Coastline Chimera, Act of Heroism
//	BlockCapacityUntilEOT{Target: t, Additional: 2}.Apply(ctx) // Yare
//	BlockCapacityUntilEOT{Target: t, AnyNumber: true}.Apply(ctx) // Valor Made Real
//	TargetCreature("…", ControlledByDefendingPlayer())           // Yare, Blaze of Glory
//	selfBlocksAtLeast(ev, source, 2)                             // Lairwatch Giant
//
// The capacity is an ADR 0041 data record (game.AddBlockCapacityMod /
// game.BlockAnyNumberMod) pinned to the creature and ending in this
// turn's cleanup step, so it survives undo and a restore point, and a
// creature that leaves and comes back is a new object it no longer
// covers (CR 400.7). It writes the same Characteristic fields the
// statics write, so it adds to them and game.BlockCapacity reads both.

// BlockCapacityUntilEOT is "<target> can block an additional N
// creatures this turn" (Additional N) or "… can block any number of
// creatures this turn" (AnyNumber). A zero Target, or neither field
// set, registers nothing.
type BlockCapacityUntilEOT struct {
	Target     uuid.UUID
	Additional int
	AnyNumber  bool
	Label      string
}

// Apply registers the record.
func (b BlockCapacityUntilEOT) Apply(ctx *Context) error {
	mods := blockCapacityMods(b.Additional, b.AnyNumber)
	if b.Target == uuid.Nil || len(mods) == 0 {
		return nil
	}
	return untilEndOfTurn(ctx, b.Target, nil, eotLabel(b.Label, "block capacity this turn"), mods...)
}

// blockCapacityMods is the mod list for a capacity grant, for a card
// that folds it into one record with a pump (Give No Ground's "+2/+6
// … and can block any number" is ONE effect, CR 613.7).
func blockCapacityMods(additional int, anyNumber bool) []game.Mod {
	switch {
	case anyNumber:
		return []game.Mod{game.BlockAnyNumberMod()}
	case additional > 0:
		return []game.Mod{game.AddBlockCapacityMod(additional)}
	}
	return nil
}

// ControlledByDefendingPlayer is "… defending player controls" on a
// spell cast during combat (#1715): a card controlled by a player
// game.IsDefendingPlayerForEffect names — under Commander's
// attack-multiple-players option every opponent of the active player,
// for the whole combat phase (CR 802.2). Outside combat nothing
// passes, so Yare cast in a main phase has no legal target.
//
// For a TRIGGER of an attacking creature ("whenever this attacks, …
// target creature defending player controls") use
// TargetCreatureDefendingPlayerControls, which reads the player that
// creature is attacking (CR 802.2a).
func ControlledByDefendingPlayer() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.IsDefendingPlayerForEffect(c.Controller)
	}
}

// selfBlocksAtLeast is "whenever this creature blocks N or more
// creatures" (Lairwatch Giant's N = 2). The lock-in emits one
// EventBlock per attacker a creature blocks and numbers them in Amount
// (#1706), so the event numbered N is the one that takes the creature
// to N — once per declaration however many more it blocks, and never
// for a creature that blocks fewer.
func selfBlocksAtLeast(ev game.Event, source *game.Card, n int) bool {
	return b28SelfBlocked(ev, source) && ev.Amount == n
}
