package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stronghold Taskmaster — Creature — Giant Minion {2}{B}{B}, 4/3 (Stronghold,
// S82 OLD b750, #5627):
//
//	"Other black creatures get -1/-1."
//
// A layer 7c modification (CR 613.4c) over every black creature but itself,
// whoever controls it, read live: a creature that turns black or stops being
// black gains or loses it at once. The Taskmaster is black but not "other", so
// it is not shrunk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a2b134d-9d7e-4d43-98fc-c776fc4bad45",
		Name:         "Stronghold Taskmaster",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			Label:    "Other black creatures get -1/-1.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.InstanceID != source.InstanceID && target.HasColorWord("B")
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power--
				c.Toughness--
			},
		}},
	})
}
