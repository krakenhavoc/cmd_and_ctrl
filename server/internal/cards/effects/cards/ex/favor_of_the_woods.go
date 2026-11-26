package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Favor of the Woods — Enchantment — Aura {2}{G} (Dark Ascension, S82 OLD,
// #5571):
//
//	"Enchant creature
//	 Whenever enchanted creature blocks, you gain 3 life."
//
// Psychic Impetus's attachment trigger on the block event: it belongs to the
// Aura, so "you" is the Aura's controller, who gains the life whoever
// controls the creature (CR 109.5). The block event names the blocker, so
// the Aura fires when its creature is declared a blocker (CR 509.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "285ba9af-bdb8-4b31-b0c5-59aa40d4e3e9",
		Name:         "Favor of the Woods",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Favor of the Woods — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
