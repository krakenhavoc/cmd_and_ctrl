package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Otherworldly Outburst — Instant {R} (Eldritch Moon, S82 OLD, #5541):
//
//	"Target creature gets +1/+0 until end of turn. When that creature dies
//	 this turn, create a 3/2 colorless Eldrazi Horror creature token."
//
// The pump, then Kelsien's delayed "when that creature dies this turn"
// about the creature as the object it is now (CR 603.7a, 603.7c),
// controlled by the caster (CR 603.7d) whoever controls the creature: the
// token is the caster's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43feb61f-5de0-41b6-b678-9785c476dc08",
		Name:         "Otherworldly Outburst",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)), game.AnswerPump),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			creature, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := effects.BoostTargetUntilEOT("Otherworldly Outburst — +1/+0", 1, 0)(ctx.Game, item); err != nil {
				return err
			}
			return effects.WhenThatCreatureDiesThisTurn(ctx, creature,
				"Otherworldly Outburst — create a 3/2 colorless Eldrazi Horror creature token", otherworldlyOutburstHorror)
		},
	})
}

var otherworldlyOutburstHorror = game.SimpleDelayedBody("otherworldly-outburst/horror", func(g *game.Game, item *game.StackItem) error {
	return effects.CreateToken{Template: effects.TokenCard("3/2 colorless Eldrazi Horror"), N: 1}.Apply(effects.NewContext(g, item))
})
