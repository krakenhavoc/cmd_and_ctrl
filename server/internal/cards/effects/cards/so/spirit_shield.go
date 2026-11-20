package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirit Shield — Artifact for {3}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {2}, {T}: Target creature gets +0/+2 for as long as this artifact remains tapped."
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
		OracleID:     "db53fad4-03b8-47bb-9459-b1eb6cca961c",
		Name:         "Spirit Shield",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Spirit Shield — you may choose not to untap this artifact")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target creature gets +0/+2 for as long as this artifact remains tapped.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetGetsWhileThisRemainsTapped("Spirit Shield — +0/+2 while tapped", game.ModifyPTMod(0, 2)),
		}},
	})
}
