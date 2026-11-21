package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mole Worms — Creature — Worm (1/1) for {2}{B}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {T}: Tap target land. It doesn't untap during its controller's untap step for as long as this creature remains tapped."
//
// ADR 0109 owner decision 4: Rust Tick's untap hold (#1313,
// TapAndHoldWhileThisRemainsTapped) with the WhileSourceRemainsTapped
// duration, on a creature. Decline to untap it and the land stays
// locked; untap it and the lock is over for good (CR 611.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b86ccfc1-a26b-4942-9d77-fef7676c77be",
		Name:         "Mole Worms",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Mole Worms — you may choose not to untap this creature")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target land. It doesn't untap during its controller's untap step for as long as this creature remains tapped.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
