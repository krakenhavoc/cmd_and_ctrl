package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resurgent Belief — Sorcery with no mana cost (Modern Horizons 2):
//
//	"Suspend 2—{1}{W} (Rather than cast this card from your hand, pay {1}{W} and
//	 exile it with two time counters on it. At the beginning of your upkeep,
//	 remove a time counter. When the last is removed, you may cast it without
//	 paying its mana cost.)
//	 Return all enchantment cards from your graveyard to the battlefield. (Auras
//	 with nothing to enchant remain in your graveyard.)"
//
// It has no mana cost, so suspend (CR 702.62a) is the whole of its playability.
// The return is Replenish's: every enchantment card of your graveyard enters
// together, Auras attaching to something they can enchant or staying where they
// are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fcfef04-a954-4423-82b0-c70dc64a8fbc",
		Name:         "Resurgent Belief",
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Suspend(2, "{1}{W}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B36ReturnAllEnchantmentCards(ctx)
		},
	})
}
