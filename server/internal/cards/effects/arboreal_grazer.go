package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arboreal Grazer — Creature — Sloth Beast {G}, 0/3 (Tarkir: Dragonstorm
// Commander, #5368):
//
//	"Reach
//	 When this creature enters, you may put a land card from your hand onto the
//	 battlefield tapped."
//
// The put is the shared clause MayPutALandFromHandTapped (Insidious Fungus,
// Embrace the Paradox): a put, not a play (CR 305.4), so the turn's land drop is
// untouched, and the land arrives tapped out of the CR 614 entry pipeline.
// Declared as a skip by an old batch only because the hand-to-battlefield move
// did not exist yet; it has since (#654).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d18a0815-59d3-4667-b52b-9acda741215e",
		Name:            "Arboreal Grazer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Arboreal Grazer — you may put a land card from your hand onto the battlefield tapped",
				Do(MayPutALandFromHandTapped("Arboreal Grazer"))),
		},
	})
}
