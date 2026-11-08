package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Consumed by Greed — Instant {1}{B}{B} (Bloomburrow, #4467):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Target opponent sacrifices a creature with the greatest power among
//	 creatures they control. If the gift was promised, return target
//	 creature card from your graveyard to your hand."
//
// The second target exists only when the gift was promised, which is
// gift.go's Instead with the printed opponent clause and the graveyard
// clause appended (Mind Spiral's shape, CR 702.174m). The gift's opponent
// and the target opponent may be different players. The edict is
// Shadowgrange Archfiend's: the target opponent chooses among their
// creatures tied for the greatest power (GreatestPowerYouControl, read
// for the chooser), and isn't asked with only one. The return is a
// second step, so it waits for that choice (CR 608.2c). Each part needs
// only its own target to be legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b548578c-4910-402c-8e78-498924e0a9be",
		Name:         "Consumed by Greed",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		Gift: GiftACard().Instead(Clauses(
			TargetPlayer("target opponent", Opponent()),
			TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
		)),
		OnResolve: ResolveSteps(
			Instruction(consumedByGreedEdict),
			Instruction(consumedByGreedReturn),
		),
	})
}

func consumedByGreedEdict(ctx *Context) error {
	who, ok := ctx.ClauseTarget(0)
	if !ok || who.Kind != game.TargetPlayer {
		return nil
	}
	ctx.Game.PlayersSacrificeForEffect(ctx.Source(), []uuid.UUID{who.ID},
		sacrificeSpec("a creature with the greatest power among creatures you control", GreatestPowerYouControl()),
		"Consumed by Greed — sacrifice a creature with the greatest power")
	return nil
}

func consumedByGreedReturn(ctx *Context) error {
	if !ctx.GiftPromised() {
		return nil
	}
	c, ok := ctx.ClauseTarget(1)
	if !ok || c.Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: c.ID, Dest: game.ZoneHand}.Apply(ctx)
}
