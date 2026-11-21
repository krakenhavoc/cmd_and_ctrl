package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lazotep Chancellor — Creature — Zombie Wizard {U}{B}, 1/3:
//
//	"Whenever you discard a card, you may pay {1}. If you do, amass
//	 Zombies 2."
//
// Bone Miser's discard condition (any card) and Mentor of the Meek's
// MayPay: one trigger per card discarded, a keyed pay branch that amasses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3e67b8f-1c91-4a2f-b924-35bbaff4bab9",
		Name:         "Lazotep Chancellor",
		Purpose:      game.Purpose{Amass: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, effects.YouDiscardedA(func(game.Card) bool { return true }),
				"Lazotep Chancellor — pay {1} to amass Zombies 2",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Lazotep Chancellor — pay {1} to amass Zombies 2?",
						Then:     lazotepChancellorAmass,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// lazotepChancellorAmass is the paid branch: amass Zombies 2. Keyed.
var lazotepChancellorAmass = effects.PayUnlessBranches("pay-unless/lazotep-chancellor-amass",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.Amass{Subtype: "Zombie", N: 2}.Apply(ctx)
	}, nil)
