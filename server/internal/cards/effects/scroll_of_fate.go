package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scroll of Fate — Artifact {3}:
//
//	"{T}: Manifest a card from your hand. (Put that card onto the
//	 battlefield face down as a 2/2 creature. Turn it face up any time
//	 for its mana cost if it's a creature card.)"
//
// You choose the card as the ability resolves. A hand of one card is
// not asked: that card is manifested. An empty hand manifests nothing.
// The card may be an instant, a sorcery, a land or a creature
// (CR 701.40a); a manifested instant or sorcery stays face down for
// good (CR 701.40g).
func init() {
	Register(Spec{
		OracleID:     "c15c13d2-8db9-4801-a1f0-f087f2551c92",
		Name:         "Scroll of Fate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: Manifest a card from your hand.",
			Cost:  TapCost(),
			// A manifest is a face-down 2/2 creature: a blocker at instant speed.
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Effect:  Do(EachManifestsFromHand{N: 1}),
		}},
	})
}
