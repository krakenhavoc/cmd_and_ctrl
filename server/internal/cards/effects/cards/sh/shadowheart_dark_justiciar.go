package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadowheart, Dark Justiciar — Legendary Creature — Human Elf Cleric
// {3}{B}, 3/4 (Commander Legends: Battle for Baldur's Gate, #5490):
//
//	"{1}{B}, {T}, Sacrifice another creature: Draw X cards, where X is
//	 that creature's power.
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// Ghoulcaller Gisa's cost with Life's Legacy's draw
// (drawEqualToSacrificedPower): X is the sacrificed creature's power as
// it last existed on the battlefield, counters and anthems included and
// floored at zero (CR 608.2h, CR 107.1b). Choose a Background is a
// deck-construction rule internal/deck reads (ADR 0144).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "180919a5-67b2-473a-9b90-a828bdc2477c",
		Name:         "Shadowheart, Dark Justiciar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}, Sacrifice another creature: Draw X cards, where X is that creature's power.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawEqualToSacrificedPower(ctx.Item, ctx)
			})),
		}},
	})
}
