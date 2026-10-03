package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rubinia Soulsinger — Legendary Creature — Faerie (2/3) for {2}{G}{W}{U}:
//
//	"You may choose not to untap Rubinia Soulsinger during your untap step.
//	 {T}: Gain control of target creature for as long as you control Rubinia Soulsinger and Rubinia Soulsinger remains tapped."
//
// ADR 0109 §3 (#1894): a control change (layer 2) for as long as you
// control Rubinia Soulsinger AND it remains tapped. That is one
// duration with two conditions (Duration.Also, CR 611.2b): it ends the
// moment either stops, so untapping Rubinia Soulsinger, or losing
// control of it, gives the creature back for good. If either is
// already false as the ability resolves, nothing is taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd3eeaba-964b-49ea-bb11-5875a78b8a4c",
		Name:         "Rubinia Soulsinger",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Rubinia Soulsinger — you may choose not to untap Rubinia Soulsinger")},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Gain control of target creature for as long as you control Rubinia Soulsinger and Rubinia Soulsinger remains tapped.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  GainControlOfTargetFor("Rubinia Soulsinger — control while you control it and it remains tapped", WhileYouControlThisAndItRemainsTapped),
		}},
	})
}
