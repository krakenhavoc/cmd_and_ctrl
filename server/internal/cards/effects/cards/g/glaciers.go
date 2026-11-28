package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glaciers — Enchantment {2}{W}{U} (Ice Age, S82 OLD batch 848, #5636):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless you pay
//	 {W}{U}.
//	 All Mountains are Plains."
//
// Aura Flux's pay-or-sacrifice upkeep tax. The static is Celestial Dawn's
// basic-land-type set (CR 305.7, layer 4) over every Mountain on the
// battlefield, whoever controls it: it becomes a Plains, loses the Mountain
// type and its abilities, and taps for {W}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4a9c2bb-e6d0-4665-8b2f-e90c78784303",
		Name:         "Glaciers",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SetsBasicLandType(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsLand() && target.HasSubtype(source.TextSubtype("Mountain")) // a word of the source's text (CR 612.2, #3984)
			}, nil, []string{"Plains"}),
		},
		Triggered: []game.TriggeredAbility{effects.OldB848UpkeepPayOrSacrifice("Glaciers", "{W}{U}")},
	})
}
