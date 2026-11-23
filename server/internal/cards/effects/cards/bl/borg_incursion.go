package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borg Incursion — Instant {2}{W} (Star Trek Commander, #4974):
//
//	"Create two 1/1 colorless Borg artifact creature tokens, then
//	 artifact creatures you control gain protection from the color of
//	 your choice until end of turn."
//
// The tokens come first, so they are among the artifact creatures
// protected. The colour is chosen as the spell resolves and asked as
// the colour to be protected FROM (ColorForProtection, #780); the set
// protected is the artifact creatures you control at that moment
// (CR 611.2c), so one that enters later this turn is not.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f00dfc16-c253-4483-94fb-3c2beabd5820",
		Name:         "Borg Incursion",
		Purpose:      game.Purpose{Answers: game.AnswerProtect | game.AnswerMakesBlocker}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Template: effects.TokenCard("1/1 colorless Borg artifact"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			then := ctx.ItemTail(effects.TrekB31ArtifactCreaturesProtectionFromChosenColor)
			then.Words = []string{"Borg Incursion"}
			effects.ChooseColorTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(),
				"Borg Incursion — choose a color", then)
			return nil
		},
	})
}
