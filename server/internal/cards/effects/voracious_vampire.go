package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voracious Vampire — Creature — Vampire Knight {2}{B}, 2/2 (Rivals of
// Ixalan, #5531):
//
//	"Menace
//	 When this creature enters, target Vampire you control gets +1/+1 and
//	 gains menace until end of turn."
//
// Menace is printed data (CR 702.111b). The entry's target is chosen as
// the trigger goes on the stack and may be the Vampire itself; the boost
// and the keyword go on together as it resolves (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c584bdf2-b328-4e50-81e3-45ca6c2aff7a",
		Name:            "Voracious Vampire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Voracious Vampire — target Vampire you control gets +1/+1 and gains menace until end of turn",
				pumpAndGrantFirstTarget(1, 1, "Voracious Vampire — +1/+1 and menace", "menace")),
				TargetCreature("target Vampire you control", YouControl(), Subtype("Vampire"))),
		},
	})
}
