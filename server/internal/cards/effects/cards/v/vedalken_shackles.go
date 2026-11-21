package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fddac567-c42c-453b-9c82-c3140b374a06",
		Name:         "Vedalken Shackles",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Vedalken Shackles — you may choose not to untap this artifact")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Gain control of target creature with power less than or equal to the number of Islands you control for as long as this artifact remains tapped.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with power less than or equal to the number of Islands you control", effects.PowerAtMostIslandsYouControl()),
			Effect:  effects.GainControlOfTargetFor("Vedalken Shackles — control while tapped", effects.DurationWhileThisRemainsTapped),
		}},
	})
}
