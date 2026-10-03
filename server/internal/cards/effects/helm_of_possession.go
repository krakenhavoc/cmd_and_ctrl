package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Helm of Possession — Artifact for {4}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {2}, {T}, Sacrifice a creature: Gain control of target creature for as long as you control this artifact and this artifact remains tapped."
//
// ADR 0109 §3 (#1894): a control change (layer 2) for as long as you
// control this artifact AND it remains tapped. That is one duration
// with two conditions (Duration.Also, CR 611.2b): it ends the moment
// either stops, so untapping this artifact, or losing control of it,
// gives the creature back for good. If either is already false as the
// ability resolves, nothing is taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78f22413-4e61-4191-9d30-7b5d6538446e",
		Name:         "Helm of Possession",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Helm of Possession — you may choose not to untap this artifact")},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a creature: Gain control of target creature for as long as you control this artifact and this artifact remains tapped.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeACreature()),
			Targets: TargetCreature("target creature"),
			Effect:  GainControlOfTargetFor("Helm of Possession — control while you control it and it remains tapped", WhileYouControlThisAndItRemainsTapped),
		}},
	})
}
