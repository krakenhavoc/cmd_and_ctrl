package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ice Floe — Land:
//
//	"You may choose not to untap this land during your untap step.
//	 {T}: Tap target creature without flying that's attacking you. It doesn't untap during its controller's untap step for as long as this land remains tapped."
//
// ADR 0109 owner decision 4: Rust Tick's untap hold (#1313,
// TapAndHoldWhileThisRemainsTapped) with the WhileSourceRemainsTapped
// duration, on a land. Decline to untap it and the creature stays
// locked; untap it and the lock is over for good (CR 611.2b).
//
// "That's attacking you" is a creature attacking Ice Floe's controller
// as a player (AttackingYou), not one attacking a planeswalker they
// control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfaaead2-09e8-47cb-9e39-8570b8d8de86",
		Name:         "Ice Floe",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Ice Floe — you may choose not to untap this land")},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target creature without flying that's attacking you. It doesn't untap during its controller's untap step for as long as this land remains tapped.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature without flying that's attacking you", WithoutKeyword("flying"), AttackingYou()),
			Effect:  TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
