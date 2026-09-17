package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gilgamesh, Master-at-Arms — Legendary Creature — Human Samurai
// {4}{R}{R}, 6/6:
//
//	"Whenever Gilgamesh enters or attacks, look at the top six cards of
//	 your library. You may put any number of Equipment cards from among
//	 them onto the battlefield. Put the rest on the bottom of your
//	 library in a random order. When you put one or more Equipment onto
//	 the battlefield this way, you may attach one of them to a Samurai
//	 you control."
//
// The shared LookAtTopThenMayPutOntoBattlefield sentence (#745) with
// "any number of Equipment cards": the six cards are looked at, only
// Gilgamesh's controller sees them, the chosen Equipment enter as one
// simultaneous batch without being cast, and the rest go to the bottom
// in an order drawn from the game's seeded RNG.
//
// DECLARED SIMPLIFICATION, weaker than printed: the reflexive "when you
// put one or more Equipment onto the battlefield this way, you may
// attach one of them to a Samurai you control" is not implemented. It
// is a reflexive triggered ability (CR 603.12) — it goes on the stack
// and can be responded to — and the engine has no way for a resolving
// ability to put a new trigger on the stack (#636). Attaching inline
// instead would deny the table the response window it prints, so the
// Equipment arrives unattached and can be equipped the ordinary way.
func init() {
	Register(Spec{
		OracleID:     "c9432e87-38f6-4889-8f07-ae7cd045e532",
		Name:         "Gilgamesh, Master-at-Arms",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Equipment put onto the battlefield aren't offered to be attached to a Samurai — equip them normally."},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Gilgamesh, Master-at-Arms — look at the top six cards; you may put any number of Equipment onto the battlefield",
				LookAtTopThenMayPutOntoBattlefield(6, OfSubtype("Equipment"), 0,
					"Gilgamesh — put any number of Equipment cards from among them onto the battlefield")),
		},
	})
}
