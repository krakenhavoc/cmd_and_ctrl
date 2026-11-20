package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// iko_b1_helpers.go — shared bodies for Ikoria: Lair of Behemoths batch 1
// (#5512). Each step acts on the spell's first card target still legal as
// it resolves, and does nothing when there is none (CR 608.2b).
// Append-only.

// ikoBoostTheTargetUntilEOT is "target creature gets +P/+T until end of
// turn" as one step of a spell's resolution.
func ikoBoostTheTargetUntilEOT(power, toughness int, label string) Instruction {
	return OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
		return BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}.Apply(ctx)
	})
}

// ikoCounterOnTheTarget is "put a <kind> counter on it" (or "on target
// creature") as one step: Fully Grown's trample counter, Unexpected
// Fangs' two counters. A keyword counter gives its keyword for as long as
// it stays (CR 122.1b).
func ikoCounterOnTheTarget(kind string) Instruction {
	return OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
		return AddCounter{Target: id, Kind: kind, N: 1}.Apply(ctx)
	})
}

// ikoUntapTheTarget is "Untap it." as the last step (Sudden Spinnerets).
func ikoUntapTheTarget() Instruction {
	return OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
		return UntapTarget{Target: id}.Apply(ctx)
	})
}

// ikoTargetGainsUntilEOT is the whole OnResolve of "Target creature gains
// <keyword> until end of turn." (Raking Claws); the grant ends at cleanup.
func ikoTargetGainsUntilEOT(keyword, label string) func(*game.StackItem, *Context) error {
	return ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
		return GrantKeywordUntilEOT{Target: id, Keywords: []string{keyword}, Label: label}.Apply(ctx)
	}))
}

// ikoSearchForABasicLandToHand is "you may search your library for a
// basic land card, reveal it, put it into your hand, then shuffle" as a
// trigger's effect (Farfinder): finding nothing is allowed (CR 701.23b).
func ikoSearchForABasicLandToHand(reason string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:    item.Controller,
			Predicate: IsBasicLand,
			Dest:      game.ZoneHand,
			Limit:     1,
			Reveal:    true,
			Shuffle:   true,
			Optional:  true,
			Reason:    reason,
		}.Apply(NewContext(g, item))
	}
}

// ikoCycledCounterOnYourCreature is "When you cycle this card, put a
// <keyword> counter on target creature you control." — Ikoria's cycle of
// keyword-counter cyclers. It triggers from the graveyard (CR 702.29c)
// and goes on the stack above the cycling ability, so the counter is put
// on before the draw; with no creature to target it is removed (CR
// 603.3d).
func ikoCycledCounterOnYourCreature(name, kind, keyword string) game.TriggeredAbility {
	return Targeting(
		InGraveyard(On(game.EventCycle, Self,
			name+" — cycled: put a "+keyword+" counter on target creature you control",
			putACounterOnTheTarget(kind))),
		TargetCreature("target creature you control", YouControl()),
	)
}
