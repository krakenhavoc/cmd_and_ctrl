package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molecule Man — Legendary Creature — Human Villain {6} 5/5:
//
//	"Nonland cards in your hand have miracle {0}. (You may cast a card for
//	 its miracle cost when you draw it if it's the first card you drew this
//	 turn.)"
//
// Miracle (CR 702.94a) granted to the cards in its controller's hand, in
// the printed shape (game/miracle.go): the static is a granted alternative
// cost (MiracleZeroForNonlandCardsInYourHand) claimable only under the hand
// permission a revealed trigger's resolution grants, and the trigger is the
// granted keyword's own, a row of this permanent that fires for the first
// card its controller drew this turn when it is a nonland card in their
// hand. Both are derived from the battlefield on every query, so the
// ability stops when Molecule Man leaves or loses its abilities. A card that
// prints a miracle of its own has both: a second instance of the keyword
// (CR 113.2c) with its own reveal, and the caster picks the cheaper.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                "ba944437-0b55-47cb-92dc-2477ce6726c3",
		Name:                    "Molecule Man",
		Completeness:            CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{MiracleZeroForNonlandCardsInYourHand()},
	})
}
