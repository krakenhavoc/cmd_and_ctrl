package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Whorl — Sorcery {2}{B} (Bloomburrow, #4467):
//
//	"Target opponent discards two cards. Then if you control a Rat,
//	 surveil 2. (Look at the top two cards of your library, then put any
//	 number of them into your graveyard and the rest on top of your
//	 library in any order.)"
//
// The discard is the opponent's own choice (CR 701.9a), queued as Mind
// Rot's is. "Then" is a second step (CR 608.2c): it waits for the discard
// to be answered, and the Rat is looked for then, on the battlefield
// through the layers and through the spell's own text (CR 612.2). The
// surveil (CR 701.25a) is the caster's prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3df4eb80-e8dd-4a9e-bd8a-b4b23fb9de6c",
		Name:         "Psychic Whorl",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.TargetPlayerDiscards(2)(ctx.Game, ctx.Item) }),
			effects.Instruction(psychicWhorlSurveil),
		),
	})
}

// psychicWhorlSurveil is "if you control a Rat, surveil 2."
func psychicWhorlSurveil(ctx *effects.Context) error {
	if !effects.ControlsA(ctx.TextSubtype("Rat"))(ctx.Game, ctx.Controller()) {
		return nil
	}
	return effects.Surveil{N: 2}.Apply(ctx)
}
