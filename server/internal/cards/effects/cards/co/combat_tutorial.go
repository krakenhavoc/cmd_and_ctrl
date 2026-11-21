package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Combat Tutorial — Sorcery {2}{U} (Final Fantasy, #4462):
//
//	"Target player draws two cards. Put a +1/+1 counter on up to one
//	 target creature you control."
//
// Cost of Brilliance's two clauses without the life loss, and the
// creature must be yours: the player (slot 0) and up to one creature
// you control (slot 1). The counter is a second step, so a draw that
// waits on a question still gets it afterwards (CR 608.2c). Each part
// reads its own slot as it resolves (CR 608.2b): a player who is no
// longer a legal target draws nothing, and a creature that is gone
// takes no counter, without stopping the other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a483d95-ff99-40b4-9071-6e78585090d5",
		Name:         "Combat Tutorial",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1),
		),
		Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2}),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(combatTutorialDraw),
			effects.Instruction(effects.CostOfBrillianceCounter),
		),
	})
}

// combatTutorialDraw is "target player draws two cards".
func combatTutorialDraw(ctx *effects.Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return effects.DrawCards{Player: t.ID, N: 2}.Apply(ctx)
}
