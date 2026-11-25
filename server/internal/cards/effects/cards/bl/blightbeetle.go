package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blightbeetle — Creature — Insect {1}{B}, 1/1 (Core Set 2020, #5518):
//
//	"Protection from green
//	 Creatures your opponents control can't have +1/+1 counters put on them."
//
// Protection from green is a printed keyword. The prohibition is a prohibition,
// not a replacement (CR 101.2): a layer 6 static sets
// Characteristic.CantHavePlusOneCounters on every creature an opponent of the
// Beetle's controller controls, and every +1/+1 counter placement, counter
// cost and entry with counters leaves those creatures out. Other kinds of
// counter are untouched, and so are the counters already on them.
//
// Seam: Characteristic.CantHavePlusOneCounters (cant_have_counters.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b4cd7e4e-4b63-41a1-bfad-2840a96f84bb",
		Name:            "Blightbeetle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from green"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			Label:     "Creatures your opponents control can't have +1/+1 counters put on them.",
			AppliesTo: effects.OldB156OpponentsCreatures,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.CantHavePlusOneCounters = true
			},
		}},
	})
}
