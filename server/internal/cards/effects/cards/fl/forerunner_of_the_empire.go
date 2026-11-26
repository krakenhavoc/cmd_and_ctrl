package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forerunner of the Empire — Creature — Human Soldier {3}{R}, 1/3:
//
//	"When this creature enters, you may search your library for a Dinosaur
//	 card, reveal it, then shuffle and put that card on top.
//	 Whenever a Dinosaur you control enters, you may have this creature deal 1
//	 damage to each creature."
//
// The first ability is the Harbinger cycle's (LrwB1Harbinger). The second
// triggers for every Dinosaur that enters under its controller's control (it is
// not a Dinosaur itself); the "you may" is the trigger's prompt, and the damage
// is dealt as one event to every creature on the battlefield, whoever controls
// it, as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "113c3024-c3bd-4c3a-8c08-3828646d9ca2",
		Name:         "Forerunner of the Empire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LrwB1Harbinger("Forerunner of the Empire", "a Dinosaur card", effects.LrwB1CardOfType("Dinosaur")),
			effects.TriggerWithPurpose(effects.Optional(effects.On(game.EventETB, effects.S82OldB216TribeEnters("Dinosaur", false),
				"Forerunner of the Empire — it deals 1 damage to each creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.Creature(), 1)
				}), "Forerunner of the Empire — deal 1 damage to each creature?"),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}}),
		},
	})
}
