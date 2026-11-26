package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace's Erasure — Enchantment {1}{U} (Magic 2012, S82 older sets batch 451,
// #5573):
//
//	"Whenever you draw a card, you may have target player mill a card."
//
// Queza's draw trigger with an optional, targeted mill of one.
//
// No simplification.
func init() {
	erasure := effects.Optional(
		effects.On(game.EventDrawCard, effects.ByYou, "Jace's Erasure — target player mills a card", effects.TargetPlayerMills(1)),
		"Jace's Erasure — have target player mill a card?")
	erasure.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{
		OracleID:     "bb7c069b-8d3f-4ba2-874b-8ec575fd6dd4",
		Name:         "Jace's Erasure",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{erasure},
	})
}
