package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Willow Satyr — Creature — Satyr (1/1) for {2}{G}{G}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {T}: Gain control of target legendary creature for as long as you control this creature and this creature remains tapped."
//
// ADR 0109 §3 (#1894): a control change (layer 2) for as long as you
// control this creature AND it remains tapped. That is one duration
// with two conditions (Duration.Also, CR 611.2b): it ends the moment
// either stops, so untapping this creature, or losing control of it,
// gives the creature back for good. If either is already false as the
// ability resolves, nothing is taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7c660bd-5f58-464c-bc75-dd244b7ca535",
		Name:         "Willow Satyr",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Willow Satyr — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Gain control of target legendary creature for as long as you control this creature and this creature remains tapped.",
			Cost:    TapCost(),
			Targets: TargetCreature("target legendary creature", Legendary()),
			Effect:  GainControlOfTargetFor("Willow Satyr — control while you control it and it remains tapped", WhileYouControlThisAndItRemainsTapped),
		}},
	})
}
