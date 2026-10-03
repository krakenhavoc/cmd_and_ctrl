package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sand Squid — Creature — Squid Beast (2/2) for {3}{U}:
//
//	"Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
//	 You may choose not to untap this creature during your untap step.
//	 {T}: Tap target creature. That creature doesn't untap during its controller's untap step for as long as this creature remains tapped."
//
// ADR 0109 owner decision 4: Rust Tick's untap hold (#1313,
// TapAndHoldWhileThisRemainsTapped) with the WhileSourceRemainsTapped
// duration, on a creature. Decline to untap it and the creature stays
// locked; untap it and the lock is over for good (CR 611.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40578245-2cb8-4a6a-aeff-2252719c2d1c",
		Name:            "Sand Squid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		UntapOptOuts:    []game.UntapOptOut{mayChooseNotToUntapSelf("Sand Squid — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target creature. That creature doesn't untap during its controller's untap step for as long as this creature remains tapped.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
