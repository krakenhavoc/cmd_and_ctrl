package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Leak — Instant {1}{U} (S82 GAP, #5948):
//
//	"Counter target spell unless its controller pays {3}."
//
// Spell Pierce's counter-unless-paid with {3} against any spell.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c61fe162-2202-4e56-9ba0-393547f9875f",
		Name:         "Mana Leak",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{3}",
				Question: "Mana Leak — pay {3} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
