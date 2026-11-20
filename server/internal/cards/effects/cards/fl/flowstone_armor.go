package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Armor — Artifact for {3}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {3}, {T}: Target creature gets +1/-1 for as long as this artifact remains tapped."
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
		OracleID:     "81f44bac-d28c-4dc1-922c-f4ebda817276",
		Name:         "Flowstone Armor",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Flowstone Armor — you may choose not to untap this artifact")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target creature gets +1/-1 for as long as this artifact remains tapped.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetGetsWhileThisRemainsTapped("Flowstone Armor — +1/-1 while tapped", game.ModifyPTMod(1, -1)),
		}},
	})
}
