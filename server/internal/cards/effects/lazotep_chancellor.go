package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c3e67b8f-1c91-4a2f-b924-35bbaff4bab9",
		Name:         "Lazotep Chancellor",
		Purpose:      game.Purpose{Amass: 2},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, youDiscardedA(func(game.Card) bool { return true }),
				"Lazotep Chancellor — pay {1} to amass Zombies 2",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Lazotep Chancellor — pay {1} to amass Zombies 2?",
						Then:     lazotepChancellorAmass,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// lazotepChancellorAmass is the paid branch: amass Zombies 2. Keyed.
var lazotepChancellorAmass = PayUnlessBranches("pay-unless/lazotep-chancellor-amass",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return Amass{Subtype: "Zombie", N: 2}.Apply(ctx)
	}, nil)
