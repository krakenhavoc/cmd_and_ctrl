package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectrum Sentinel — Artifact Creature — Soldier {1}, 1/2 (The Brothers'
// War):
//
//	"Protection from multicolored
//	 Whenever a nonbasic land an opponent controls enters, you gain 1
//	 life."
//
// Protection from multicolored is Stonecoil Serpent's single protection
// ability (game.ProtectionQualityMulticolored). The trigger is Polluted
// Bonds' "land an opponent controls enters" watch narrowed to lands
// without the Basic supertype; the 1 life is gained by Spectrum
// Sentinel's controller whether or not the opponent is still seated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7c13d0ad-1235-41f7-8f9a-f39d7c544b74",
		Name:            "Spectrum Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from multicolored"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				land, ok := effects.B34LandAnOpponentControlsEntered(ev, source, g)
				return ok && !land.HasSupertype("Basic")
			},
			Key: "Spectrum Sentinel — you gain 1 life",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Spectrum Sentinel — you gain 1 life")
			},
			Effect: effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
}
