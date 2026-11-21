package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crash of Rhino Beetles — Creature — Insect {4}{G}, 5/5 (Commander 2018,
// #5526):
//
//	"Trample
//	 This creature gets +10/+10 as long as you control ten or more lands."
//
// The pump is a layer 7c modification of the Beetles alone (CR 613.4c),
// read on every recompute (CR 611.3a): Fecund Greenshell's ten-land count,
// which moves with the battlefield, and the battlefield already drops the
// layer cache.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "61dd63b7-11dc-49db-8106-36d455a182be",
		Name:            "Crash of Rhino Beetles",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:    "This creature gets +10/+10 as long as you control ten or more lands.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: effects.WhileYouHave(func(g *game.Game, you uuid.UUID) bool {
				return effects.B02CountLandsControlledBy(g, you) >= 10
			}, effects.SelfCreature),
			Apply: effects.PumpBy(10, 10),
		}},
	})
}
