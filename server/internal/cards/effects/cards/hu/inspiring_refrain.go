package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiring Refrain — Sorcery {4}{U}{U} (Doctor Who):
//
//	"Draw two cards. Exile Inspiring Refrain with three time counters on it.
//	 Suspend 3—{2}{U} (Rather than cast this card from your hand, you may pay
//	 {2}{U} and exile it with three time counters on it. At the beginning of
//	 your upkeep, remove a time counter. When the last is removed, you may cast
//	 it without paying its mana cost.)"
//
// Rousing Refrain's shape: the draw, then the spell exiles itself with three
// time counters (so it suspends and, being a suspend card, comes back at
// the last counter, and exiles itself again when it resolves). Suspend is the
// engine's special action (CR 702.62).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "1773c4e8-dc08-482a-9f26-30ef96c5b200",
		Name:           "Inspiring Refrain",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 2},
		SpecialActions: []game.SpecialAction{effects.Suspend(3, "{2}{U}")},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 2},
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.Cmdr2S40ExileWithTimeCounters(ctx, 3)
			}),
		),
	})
}
