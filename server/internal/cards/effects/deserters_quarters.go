package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deserter's Quarters — Artifact for {2}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {6}, {T}: Tap target creature. It doesn't untap during its controller's untap step for as long as this artifact remains tapped."
//
// ADR 0109 owner decision 4: Rust Tick's untap hold (#1313,
// TapAndHoldWhileThisRemainsTapped) with the WhileSourceRemainsTapped
// duration, on a artifact. Decline to untap it and the creature stays
// locked; untap it and the lock is over for good (CR 611.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa8e0e77-4bbe-42d3-a3db-034fd114f9d7",
		Name:         "Deserter's Quarters",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Deserter's Quarters — you may choose not to untap this artifact")},
		Activated: []ActivatedAbility{{
			Label:   "{6}, {T}: Tap target creature. It doesn't untap during its controller's untap step for as long as this artifact remains tapped.",
			Cost:    Plus(ManaCost("{6}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
