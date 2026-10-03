package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Gremlins — Creature — Phyrexian Gremlin (1/1) for {2}{B}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {T}: Tap target artifact. It doesn't untap during its controller's untap step for as long as this creature remains tapped."
//
// ADR 0109 owner decision 4: Rust Tick's untap hold (#1313,
// TapAndHoldWhileThisRemainsTapped) with the WhileSourceRemainsTapped
// duration, on a creature. Decline to untap it and the artifact stays
// locked; untap it and the lock is over for good (CR 611.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "407a0761-7ccc-4607-8df6-e744d30a81a0",
		Name:         "Phyrexian Gremlins",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Phyrexian Gremlins — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target artifact. It doesn't untap during its controller's untap step for as long as this creature remains tapped.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
