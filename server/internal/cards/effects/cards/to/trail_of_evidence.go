package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trail of Evidence — Enchantment {2}{U} (Shadows over Innistrad, #5543):
//
//	"Whenever you cast an instant or sorcery spell, investigate."
//
// Balmor's cast condition (an instant or a sorcery you cast), with the Clue as
// the effect. It triggers on the cast, before the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e10065ba-fe24-4974-a500-d4151397a4d9",
		Name:         "Trail of Evidence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Trail of Evidence — investigate", effects.Do(effects.Investigate{N: 1})),
		},
	})
}
