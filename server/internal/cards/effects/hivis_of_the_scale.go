package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hivis of the Scale — Legendary Creature — Lizard Shaman (3/4) for {3}{R}{R}:
//
//	"You may choose not to untap Hivis during your untap step.
//	 {T}: Gain control of target Dragon for as long as you control Hivis and Hivis remains tapped."
//
// ADR 0109 §3 (#1894): a control change (layer 2) for as long as you
// control Hivis AND it remains tapped. That is one duration with two
// conditions (Duration.Also, CR 611.2b): it ends the moment either
// stops, so untapping Hivis, or losing control of it, gives the
// creature back for good. If either is already false as the ability
// resolves, nothing is taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "797adc7e-03a9-4221-bdaf-4c46dcb95701",
		Name:         "Hivis of the Scale",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Hivis of the Scale — you may choose not to untap Hivis")},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Gain control of target Dragon for as long as you control Hivis and Hivis remains tapped.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target Dragon", HasSubtype("Dragon")),
			Effect:  GainControlOfTargetFor("Hivis of the Scale — control while you control it and it remains tapped", WhileYouControlThisAndItRemainsTapped),
		}},
	})
}
