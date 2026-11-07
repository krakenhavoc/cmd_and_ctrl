package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7e074b14-0ec8-439b-9b83-5f7c1a904384",
		Name:         "Haunt the Network",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 2},
			// The drain waits for any prompt the tokens ask (CR 608.2c).
			Instruction(hauntTheNetworkDrain),
		),
	})
}

// hauntTheNetworkDrain is "the chosen player loses X life and you gain X
// life", X the artifacts the caster controls as it runs.
func hauntTheNetworkDrain(ctx *Context) error {
	x := len(permanentsControlledByMatching(ctx.Game, ctx.Controller(), Artifact()))
	if x == 0 {
		return nil
	}
	if err := (TargetPlayerLosesLife{N: x}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: ctx.Controller(), Amount: x}.Apply(ctx)
}
