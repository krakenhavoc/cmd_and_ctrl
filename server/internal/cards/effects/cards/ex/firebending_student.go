package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firebending Student — Creature — Human Monk {1}{R}, 1/2 (Avatar: The
// Last Airbender, #4459):
//
//	"Prowess
//	 Firebending X, where X is this creature's power."
//
// X is the Student's power as the firebending trigger resolves, or its
// last known power if it has left (CR 608.2h), so a prowess pump cast in
// response counts. Prowess is the engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1ad3be51-f68b-46d6-b58e-400b50e384f1",
		Name:            "Firebending Student",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Firebending Student — firebending X, where X is its power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					// X of zero adds nothing but still firebends (#5276).
					return effects.NewFirebending(effects.SourcePowerNow(ctx)).Apply(ctx)
				}),
		},
	})
}
