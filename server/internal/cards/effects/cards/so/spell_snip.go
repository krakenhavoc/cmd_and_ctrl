package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spell Snip — Instant {2}{U} (Shards of Alara, S82 OLD batch 509, #5588):
//
//	"Counter target spell unless its controller pays {1}.
//	 Cycling {2}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfdae813-7241-4f6c-99f0-c772af34af91",
		Name:         "Spell Snip",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{1}",
				Question: "Spell Snip — pay {1} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
