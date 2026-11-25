package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaervek, the Spiteful — Legendary Creature — Human Warlock {2}{B}{B}, 3/2
// (Core Set 2021, #5511):
//
//	"Other creatures get -1/-1."
//
// A layer 7c static over every creature but Kaervek, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b338d879-189d-4b74-9d63-42920fdb9855",
		Name:         "Kaervek, the Spiteful",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Other creatures get -1/-1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.InstanceID != source.InstanceID
			},
			Apply: effects.PumpBy(-1, -1),
		}},
	})
}
