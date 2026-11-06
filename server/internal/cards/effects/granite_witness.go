package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Granite Witness — Artifact Creature — Gargoyle Detective {2}{W}{U}, 3/2
// (Murders at Karlov Manor, #4469):
//
//	"Flying, vigilance
//	 Disguise {W/U}{W/U} (You may cast this card face down for {3} as a
//	 2/2 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this creature is turned face up, you may tap or untap target
//	 creature."
//
// The face-up trigger (CR 708.8) is Sewer-veillance Cam's "you may tap or
// untap target creature": the target is chosen as the trigger goes on the
// stack, the "you may" is asked then, and tap or untap is picked as it
// resolves (b38TapOrUntapTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "37f2ec12-d703-40cd-a155-053ae2e4ddcc",
		Name:             "Granite Witness",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying", "vigilance"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{W/U}{W/U}")},
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					WhenThisIsTurnedFaceUp("Granite Witness — tap or untap target creature",
						b38TapOrUntapTarget("Granite Witness")),
					TargetCreature("target creature")),
				"Granite Witness — tap or untap target creature?"),
		},
	})
}
