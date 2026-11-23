package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "40578245-2cb8-4a6a-aeff-2252719c2d1c",
		Name:            "Sand Squid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		UntapOptOuts:    []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Sand Squid — you may choose not to untap this creature")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target creature. That creature doesn't untap during its controller's untap step for as long as this creature remains tapped.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
