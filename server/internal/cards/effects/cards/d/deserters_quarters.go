package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "aa8e0e77-4bbe-42d3-a3db-034fd114f9d7",
		Name:         "Deserter's Quarters",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Deserter's Quarters — you may choose not to untap this artifact")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}, {T}: Tap target creature. It doesn't untap during its controller's untap step for as long as this artifact remains tapped.",
			Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
