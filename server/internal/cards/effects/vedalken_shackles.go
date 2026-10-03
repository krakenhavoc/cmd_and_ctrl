package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vedalken Shackles — Artifact for {3}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {2}, {T}: Gain control of target creature with power less than or equal to the number of Islands you control for as long as this artifact remains tapped."
//
// ADR 0109 owner decision 4: a control change (layer 2) for as long as
// the Shackles remain tapped (CR 611.2b). The count of Islands is read
// as the target is chosen and again as the ability resolves (CR
// 608.2b); a creature whose power has grown past it by then is an
// illegal target and nothing is taken. Once taken, the creature is
// kept however the count changes: the text bounds the target, not the
// duration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fddac567-c42c-453b-9c82-c3140b374a06",
		Name:         "Vedalken Shackles",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Vedalken Shackles — you may choose not to untap this artifact")},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Gain control of target creature with power less than or equal to the number of Islands you control for as long as this artifact remains tapped.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target creature with power less than or equal to the number of Islands you control", PowerAtMostIslandsYouControl()),
			Effect:  GainControlOfTargetFor("Vedalken Shackles — control while tapped", DurationWhileThisRemainsTapped),
		}},
	})
}
