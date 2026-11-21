package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peter Parker's Camera — Artifact {1} (Marvel's Spider-Man):
//
//	"This artifact enters with three film counters on it.
//	 {2}, {T}, Remove a film counter from this artifact: Copy target
//	 activated or triggered ability you control. You may choose new
//	 targets for the copy."
//
// Lithoform Engine's first ability (AbilityOnStack + copyTargetedAbility)
// with the counter paid as a cost at announcement (Braided Net's
// RemoveCountersFromThis), so a response cannot spend it twice. The
// counters come in as a CR 614.1c "enters with".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10777360-c046-43e3-ab44-9d0b926fbbf8",
		Name:         "Peter Parker's Camera",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("film", 3, "Peter Parker's Camera: enters with three film counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Remove a film counter from this artifact: Copy target activated or triggered ability you control. You may choose new targets for the copy.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.RemoveCountersFromThis("film", 1)),
			Targets: effects.AbilityOnStack("target activated or triggered ability you control", effects.AnAbilityYouControl()),
			Effect:  effects.CopyTargetedAbility,
		}},
	})
}
