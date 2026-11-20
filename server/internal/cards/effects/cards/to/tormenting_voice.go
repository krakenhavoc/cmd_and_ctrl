package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tormenting Voice — Sorcery {1}{R}:
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards."
//
// Thrill of Possibility's printed twin, one mana value earlier in the
// format's history. Same reasoning: the discard is part of casting,
// so a discard payoff (Mary Read and Anne Bonny, Marauding Mako) sees
// it with the spell already on the stack, and countering the spell
// doesn't give the card back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "f307b5b4-e949-4f69-8dc7-856e33a45a16",
		Name:           "Tormenting Voice",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
