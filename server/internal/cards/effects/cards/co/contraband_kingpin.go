package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contraband Kingpin — Creature — Aetherborn Rogue {U}{B}, 1/4 (Kaladesh,
// #5539):
//
//	"Lifelink
//	 Whenever an artifact you control enters, scry 1."
//
// Lifelink is printed. The trigger is Weldfast Wingsmith's "artifact entered
// under your control" watcher with Biblioplex Kraken's scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "680da572-6cd5-482a-b01d-32944566d8c4",
		Name:            "Contraband Kingpin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.IsArtifactEnteredUnderYourControl(ev, source, g)
			}, "Contraband Kingpin — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
