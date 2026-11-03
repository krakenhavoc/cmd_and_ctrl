package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zaffai and the Tempests — Legendary Creature — Human Bard Sorcerer
// {5}{U}{R}, 5/7 (Secrets of Strixhaven Commander, #4834):
//
//	"Once during each of your turns, you may cast an instant or sorcery
//	 spell from your hand without paying its mana cost."
//
// A standing granted alternative cost (ADR 0118 §3, CR 118.9): free, from
// the hand only, for instant and sorcery spells only, spent once per
// Zaffai object as the spell becomes cast, and offered only during its
// controller's own turn (soc_b18_helpers.go). Casting without paying the
// mana cost still obeys the card's own timing: a sorcery needs a main
// phase and an empty stack, an instant any time its owner has priority on
// their turn. A spell with {X} in its cost is cast with X = 0 (the ruling,
// CR 107.3b); additional costs are still paid (CR 118.9d), and a failed or
// refused cast spends nothing. A second Zaffai opens another use.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f8a8650-0e38-4926-8f48-526351e6a863",
		Name:         "Zaffai and the Tempests",
		Completeness: CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{
			CastAnInstantOrSorceryFromHandFreeOnceEachOfYourTurns(),
		},
	})
}
