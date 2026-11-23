package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cartel Aristocrat — Creature — Human Advisor {W}{B}, 2/2 (Ravnica
// Remastered, #5948):
//
//	"Sacrifice another creature: This creature gains protection from the
//	 color of your choice until end of turn."
//
// "Another" is object identity (SacrificeAnotherN). The colour is chosen as
// the ability resolves, not when it is activated, by its controller
// (chooseProtectionColorFor, Akroma's Blessing's keyed tail), and the grant
// lasts until end of turn on the Aristocrat as it is now.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9603e553-f72d-40bf-94ab-596afb7d6db3",
		Name:         "Cartel Aristocrat",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: This creature gains protection from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerProtect},
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				effects.ChooseProtectionColorFor(ctx, ctx.Controller(), item.SourceCardID, "Cartel Aristocrat",
					"Cartel Aristocrat — choose a color to gain protection from")
				return nil
			},
		}},
	})
}
