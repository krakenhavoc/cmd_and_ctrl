package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zuko, Exiled Prince — Legendary Creature — Human Noble {3}{R}, 4/3
// (Avatar: The Last Airbender, #4459):
//
//	"Firebending 3 (Whenever this creature attacks, add {R}{R}{R}. This
//	 mana lasts until end of combat.)
//	 {3}: Exile the top card of your library. You may play that card
//	 this turn."
//
// Firebending is Avatar Roku's attack trigger with N = 3, a triggered
// ability that uses the stack, its mana marked to last until end of
// combat (the rulings). The {3} ability is Hedron Detonator's impulse:
// the exiled card may be played this turn under the normal timing rules
// and for its costs (the third ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3231ce2e-d4f8-44fc-872d-aeafa59aed4b",
		Name:         "Zuko, Exiled Prince",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Zuko, Exiled Prince — firebending 3", Do(firebending(3))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}: Exile the top card of your library. You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				_, err := b12ImpulseExileForTurn(g, item, 1)
				return err
			},
		}},
	})
}
