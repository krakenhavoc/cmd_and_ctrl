package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashnod's Battle Gear — Artifact for {2}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {2}, {T}: Target creature you control gets +2/-2 for as long as this artifact remains tapped."
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
		OracleID:     "b5a390fd-2864-4481-84b4-41e8fac91a80",
		Name:         "Ashnod's Battle Gear",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Ashnod's Battle Gear — you may choose not to untap this artifact")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target creature you control gets +2/-2 for as long as this artifact remains tapped.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  effects.TargetGetsWhileThisRemainsTapped("Ashnod's Battle Gear — +2/-2 while tapped", game.ModifyPTMod(2, -2)),
		}},
	})
}
