package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "3fdc082b-1abf-4715-9d66-105981c1f989",
		Name:            "Steelgaze Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Steelgaze Griffin — +2/+0 until end of turn",
				thisCreatureUntilEOT("Steelgaze Griffin — +2/+0", 2, 0)),
		},
	})
}
