package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lunk Errant — Creature — Giant Warrior {5}{R}, 4/4 (Morningtide, #5592):
//
//	"Whenever this creature attacks alone, it gets +1/+1 and gains trample
//	 until end of turn."
//
// Intrepid Trufflesnout's "attacks alone" (ThisAttackedAlone, CR 506.5:
// the only creature declared as an attacker), with the self pump and the
// self trample grant, both ending at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c2445b4-e8be-452c-950d-f42e549daf9a",
		Name:         "Lunk Errant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttackedAlone, "Lunk Errant — +1/+1 and trample until end of turn", lunkErrantPump),
		},
	})
}

// lunkErrantPump is "it gets +1/+1 and gains trample until end of turn".
func lunkErrantPump(g *game.Game, item *game.StackItem) error {
	if err := thisGetsUntilEndOfTurn(1, 1, "Lunk Errant — +1/+1")(g, item); err != nil {
		return err
	}
	return thisGainsKeywordUntilEndOfTurn("trample", "Lunk Errant — trample")(g, item)
}
