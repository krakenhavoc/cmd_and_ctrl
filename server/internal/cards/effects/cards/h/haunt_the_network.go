package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haunt the Network — Sorcery {3}{U}{B} (Aetherdrift, #4464):
//
//	"Choose target opponent. Create two 1/1 colorless Thopter artifact
//	 creature tokens with flying. Then the chosen player loses X life and
//	 you gain X life, where X is the number of artifacts you control."
//
// The opponent is the spell's one target, so an opponent gone by
// resolution (one who has left the game) counters it and there are no
// Thopters (CR 608.2b). The tokens are the catalog's predefined Thopter.
// "Then" makes the drain wait for the tokens, so X is counted as the drain
// runs and the two new Thopters count; an artifact Vehicle counts whether
// or not it is crewed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e074b14-0ec8-439b-9b83-5f7c1a904384",
		Name:         "Haunt the Network",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 2},
			// The drain waits for any prompt the tokens ask (CR 608.2c).
			effects.Instruction(hauntTheNetworkDrain),
		),
	})
}

// hauntTheNetworkDrain is "the chosen player loses X life and you gain X
// life", X the artifacts the caster controls as it runs.
func hauntTheNetworkDrain(ctx *effects.Context) error {
	x := len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Artifact()))
	if x == 0 {
		return nil
	}
	if err := (effects.TargetPlayerLosesLife{N: x}).Apply(ctx); err != nil {
		return err
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: x}.Apply(ctx)
}
