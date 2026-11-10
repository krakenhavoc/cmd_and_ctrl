package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Axebane Ferox — Creature — Beast {2}{G}{G} (Murders at Karlov Manor, #4469):
//
//	"Deathtouch, haste
//	 Ward—Collect evidence 4. (Whenever this creature becomes the target of a
//	 spell or ability an opponent controls, counter it unless that player
//	 exiles cards with total mana value 4 or greater from their graveyard.)"
//
// The ward is the collect evidence seam (WardCollectEvidence,
// std_b43_ward_evidence.go): the payer is the player who targeted it, and a
// graveyard with less than 4 in all can't pay (CR 701.59b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "35b3ae1c-4116-49b0-89a9-7332fe9bbff0",
		Name:            "Axebane Ferox",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "haste"},
		Triggered: []game.TriggeredAbility{
			Ward(WardCollectEvidence(4), "Axebane Ferox — ward—collect evidence 4"),
		},
	})
}
