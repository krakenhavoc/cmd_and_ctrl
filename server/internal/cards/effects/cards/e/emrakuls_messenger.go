package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emrakul's Messenger — Creature — Eldrazi Faerie Rogue {1}{U}, 2/1
// (Modern Horizons 3):
//
//	"Devoid (This card has no color.)
//	 Flying
//	 Whenever you draw your second card each turn, create a 0/1
//	 colorless Eldrazi Spawn creature token with "Sacrifice this token:
//	 Add {C}.""
//
// Atlantean Cavalry's trigger (b41DrewYourSecondCardThisTurn reads the
// turn's draw tally after the draw, so it fires on the second card and no
// other) with Chittering Dispatcher's Eldrazi Spawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f1409c9e-f0d9-49e1-b02e-d0f04a658331",
		Name:            "Emrakul's Messenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Emrakul's Messenger — create a 0/1 colorless Eldrazi Spawn",
				effects.Do(effects.CreateToken{Template: effects.EldraziSpawnToken(), N: 1})),
		},
	})
}
