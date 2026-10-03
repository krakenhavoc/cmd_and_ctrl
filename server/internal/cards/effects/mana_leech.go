package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cad05528-917b-4296-9d23-ba9433612709",
		Name:         "Mana Leech",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Mana Leech — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target land. It doesn't untap during its controller's untap step for as long as this creature remains tapped.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
