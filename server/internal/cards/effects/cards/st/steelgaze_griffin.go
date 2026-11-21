package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steelgaze Griffin — Creature — Griffin {4}{U}, 2/4 (Throne of Eldraine, #5516):
//
//	"Flying
//	 Whenever you draw your second card each turn, this creature gets +2/+0 until end of turn."
//
// Thopter Fabricator's "second card each turn" condition
// (b41DrewYourSecondCardThisTurn): it fires once a turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3fdc082b-1abf-4715-9d66-105981c1f989",
		Name:            "Steelgaze Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Steelgaze Griffin — +2/+0 until end of turn",
				effects.ThisCreatureUntilEOT("Steelgaze Griffin — +2/+0", 2, 0)),
		},
	})
}
