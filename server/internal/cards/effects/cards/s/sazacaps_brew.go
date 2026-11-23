package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sazacap's Brew — Instant {1}{R} (Bloomburrow, #4467):
//
//	"Gift a tapped Fish (You may promise an opponent a gift as you cast
//	 this spell. If you do, they create a tapped 1/1 blue Fish creature
//	 token before its other effects.)
//	 As an additional cost to cast this spell, discard a card.
//	 Target player draws two cards. If the gift was promised, target
//	 creature you control gets +2/+0 until end of turn."
//
// Mind Spiral's gift: the creature clause exists only when the gift was
// promised (CR 702.174m), so the promise gives the spell a second clause
// through Instead, and the Fish comes first (CR 702.174j). The discard
// is Big Score's additional cost (CR 601.2b), paid whether or not the
// gift is promised. Each clause is read on its own at resolution: a
// player who left still lets the creature get +2/+0, and a creature that
// left still lets the player draw (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ac7d1986-8388-4e4c-973e-d90382b12acf",
		Name:           "Sazacap's Brew",
		Completeness:   effects.CompletenessFull,
		Purpose:        effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2}),
		AdditionalCost: effects.DiscardCost(1),
		Targets:        effects.TargetPlayer("target player"),
		Gift: effects.GiftATappedFish().Instead(effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.TargetCreature("target creature you control", effects.YouControl()),
		)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(sazacapsBrewDraw),
			effects.Instruction(sazacapsBrewPump),
		),
	})
}

func sazacapsBrewDraw(ctx *effects.Context) error {
	who, ok := ctx.ClauseTarget(0)
	if !ok || who.Kind != game.TargetPlayer {
		return nil
	}
	return effects.DrawCards{Player: who.ID, N: 2}.Apply(ctx)
}

func sazacapsBrewPump(ctx *effects.Context) error {
	if !ctx.GiftPromised() {
		return nil
	}
	c, ok := ctx.ClauseTarget(1)
	if !ok || c.Kind != game.TargetCard {
		return nil
	}
	return effects.BoostUntilEOT{Target: c.ID, Power: 2, Label: "Sazacap's Brew — +2/+0 until end of turn"}.Apply(ctx)
}
