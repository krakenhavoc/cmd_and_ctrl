package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Escaped Null — Creature — Zombie Berserker {3}{B}, 1/2 (Rise of the
// Eldrazi, S82 OLD batch 483, #5581):
//
//	"Lifelink
//	 Whenever this creature blocks or becomes blocked, it gets +5/+0 until end
//	 of turn."
//
// Bushido's trigger shape (CR 509.3a, 509.3c): once however many creatures
// it blocks or is blocked by, and it is the Null itself that gets the boost.
//
// No simplification.
func init() {
	const label = "Escaped Null — +5/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "8bc17c50-f4c9-46dd-9294-3036ed46de91",
		Name:            "Escaped Null",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.ThisBlocksOrBecomesBlocked,
				label, effects.ThisGetsUntilEndOfTurn(5, 0, label)),
		},
	})
}
