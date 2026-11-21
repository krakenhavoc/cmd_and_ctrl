package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endoskeleton — Artifact for {2}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {2}, {T}: Target creature gets +0/+3 for as long as this artifact remains tapped."
//
// ADR 0109 owner decision 4: a continuous effect from a resolving
// ability (CR 611.2), timed by "for as long as this artifact remains
// tapped" (CR 611.2b). The effect is a ScopedEffect record, data a
// restore point carries; it ends the moment the artifact untaps, and
// for good. "You may choose not to untap" is the untap opt-out, which
// is how its controller keeps the effect going.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9ef7f8c-21ef-45f0-af64-8ffa9f0089d1",
		Name:         "Endoskeleton",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Endoskeleton — you may choose not to untap this artifact")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target creature gets +0/+3 for as long as this artifact remains tapped.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetGetsWhileThisRemainsTapped("Endoskeleton — +0/+3 while tapped", game.ModifyPTMod(0, 3)),
		}},
	})
}
