package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Q, Your Illustrious Host — {W}{U}{B}{R}{G} Legendary Creature — Q 5/5
// (Star Trek Commander, #4974):
//
//	"Tap three untapped Faeries you control: Exile cards from the top of
//	 your library until you exile a nonland card. You may cast that card
//	 without paying its mana cost."
//
// The cost is High Perfect Morcant's tap-three (TapOthersCost, three
// untapped Faerie creatures you control): not the {T} symbol, so Faeries
// that arrived this turn may pay it (CR 302.6), and it does not target.
// Q is not a Faerie and has no {T}, so the ability can be activated again
// with three more. The effect is Nico Minoru, Runaway's
// exileUntilNonlandCastFree: the walk, then a cast during the resolution
// at {0} (an {X} is 0, CR 107.3b); a card not cast stays in exile, and so
// do the lands exiled on the way, as the card says nothing else.
//
// No simplification.
func init() {
	const tapThree = "three untapped Faeries you control"
	Register(Spec{
		OracleID:     "e4e96682-1130-49cc-b1e5-04a774024f7a",
		Name:         "Q, Your Illustrious Host",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap three untapped Faeries you control: Exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: TargetPermanent(tapThree, Creature(), OfCreatureType("Faerie")),
				Label:  tapThree,
			}},
			Effect: qYourIllustriousHostExile,
		}},
	})
}

func qYourIllustriousHostExile(g *game.Game, item *game.StackItem) error {
	return exileUntilNonlandCastFree(NewContext(g, item), "Q, Your Illustrious Host")
}
