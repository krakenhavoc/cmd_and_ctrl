package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noble Stand — Enchantment {4}{W} (Nemesis, S82 OLD batch 680, #5618):
//
//	"Whenever a creature you control blocks, you gain 2 life."
//
// CR 509.3a: one trigger per blocking creature, however many attackers it
// blocks (EventBlock carries a per-pair count; only the first reads).
// "You control" is the blocker's controller as it blocks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1aa1e34c-b2ff-4750-9e73-5707a8f49bfb",
		Name:         "Noble Stand",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 2},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBlock && ev.Actor == source.Controller && ev.Amount <= 1
			}, "Noble Stand — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
