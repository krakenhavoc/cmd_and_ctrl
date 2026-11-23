package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mana Leech — Creature — Leech (1/1) for {2}{B}:
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
		OracleID:     "cad05528-917b-4296-9d23-ba9433612709",
		Name:         "Mana Leech",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Mana Leech — you may choose not to untap this creature")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target land. It doesn't untap during its controller's untap step for as long as this creature remains tapped.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
