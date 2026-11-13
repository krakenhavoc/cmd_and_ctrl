package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampire Hounds — Creature — Vampire Dog {2}{B}, 2/2 (Exodus, #5626):
//
//	"Discard a creature card: This creature gets +2/+2 until end of
//	 turn."
//
// Floral Evoker's "Discard a creature card" cost (DiscardCardsMatching
// with MatchCreature, paid as it is activated, CR 602.2b) on the shared
// thisGetsUntilEndOfTurn body. A noncreature card can't pay it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0184d74b-e9aa-4b5a-8db0-b15cecbc17fb",
		Name:         "Vampire Hounds",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a creature card: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    DiscardCardsMatching(1, "a creature card", MatchCreature),
			Effect:  thisGetsUntilEndOfTurn(2, 2, "Vampire Hounds — +2/+2 until end of turn"),
		}},
	})
}
