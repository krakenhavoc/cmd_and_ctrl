package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overgrown Armasaur — Creature — Dinosaur {3}{G}{G}, 4/4 (Rivals of
// Ixalan, #5531):
//
//	"Enrage — Whenever this creature is dealt damage, create a 1/1 green
//	 Saproling creature token."
//
// The shared enrage trigger (Ripjaw Raptor's shape) with a Saproling from
// the token table; lethal damage still makes the token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e104718-1f74-4414-977a-82def93b5a8e",
		Name:         "Overgrown Armasaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Overgrown Armasaur — Enrage: create a 1/1 green Saproling",
				Do(CreateToken{Template: b11GreenSaprolingToken(), N: 1})),
		},
	})
}
