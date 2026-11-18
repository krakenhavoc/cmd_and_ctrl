package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whispering Wizard — Creature — Human Wizard {3}{U}, 3/2 (Innistrad:
// Crimson Vow):
//
//	"Whenever you cast a noncreature spell, create a 1/1 white Spirit
//	 creature token with flying. This ability triggers only once each
//	 turn."
//
// The trigger is G'raha Tia's cast condition with TriggersOnlyOnceEachTurn:
// the second noncreature spell of a turn does not trigger it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "321e101c-a426-4d18-be0b-c26c5348b70d",
		Name:         "Whispering Wizard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(On(game.EventCast, YouCast(Noncreature()),
				"Whispering Wizard — create a 1/1 white Spirit creature token with flying",
				Do(vowB12SpiritToken()))),
		},
	})
}
