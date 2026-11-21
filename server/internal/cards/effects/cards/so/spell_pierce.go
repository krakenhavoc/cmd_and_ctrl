package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spell Pierce — Instant {U}:
//
//	"Counter target noncreature spell unless its controller pays {2}."
//
// Flusterstorm's pay-unless on a wider target: the prompt goes to the
// spell's controller, and the spell cannot resolve while the {2} is
// outstanding. Being a pay-unless it is also legal against a spell
// that can't be countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d64b0848-0193-4025-ba62-63ecd8fb9f50",
		Name:         "Spell Pierce",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target noncreature spell", effects.Noncreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{2}",
				Question: "Spell Pierce — pay {2} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
