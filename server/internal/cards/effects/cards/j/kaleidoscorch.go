package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaleidoscorch — Sorcery {1}{R} (Modern Horizons 2, #5501):
//
//	"Converge — Kaleidoscorch deals X damage to any target, where X is the
//	 number of colors of mana spent to cast this spell.
//	 Flashback {4}{R}"
//
// X is the count of distinct colours of the mana paid for this cast, so a
// flashback cast counts the colours of the {4}{R} it paid (Prismatic
// Ending's read).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "b02bd695-ec7f-47be-b797-e2b4b309cc2a",
		Name:                "Kaleidoscorch",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Targets:             effects.TargetAny(),
		CastableZones:       []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts:    []game.AlternativeCost{effects.Flashback("{4}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DealDamageToTheTarget(ctx.ColorsSpentCount())(ctx.Game, item)
		},
	})
}
