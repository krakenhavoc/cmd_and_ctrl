package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ashiok's Adept — Creature — Human Wizard {2}{B}, 1/3 (Born of the Gods,
// #5559):
//
//	"Heroic — Whenever you cast a spell that targets this creature,
//	 each opponent discards a card."
//
// The heroic seam (heroic) with Burglar Rat's body
// (eachOpponentDiscardsOne): each opponent picks the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80eca17e-5dfd-4051-9f94-0c8aaf3747bf",
		Name:         "Ashiok's Adept",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			heroic("Ashiok's Adept — each opponent discards a card (heroic)", eachOpponentDiscardsOne),
		},
	})
}
