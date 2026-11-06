package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruin-Lurker Bat — Creature — Bat {W}, 1/1 (The Lost Caverns of Ixalan,
// #4470):
//
//	"Flying, lifelink
//	 At the beginning of your end step, if you descended this turn,
//	 scry 1."
//
// Flying and lifelink are printed keywords (CR 702.9a, CR 702.15a). The
// end-step trigger is Deep Goblin Skulltaker's AtYourEndStepIfDescended:
// "you descended this turn" (CR 700.11) is an intervening if, checked as
// the step begins and again on resolution (CR 603.4), so a turn with no
// permanent card put into its controller's graveyard puts nothing on the
// stack. The scry is the footprinted primitive (CR 701.22a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e6f379ad-96b1-46c3-bbc6-a8cb57156a7f",
		Name:            "Ruin-Lurker Bat",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Scry: 1},
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfDescended("Ruin-Lurker Bat — scry 1", Do(Scry{N: 1})),
		},
	})
}
