package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonlair Spider — Creature — Spider {2}{R}{R}{G}{G}, 5/6 (Conspiracy: Take
// the Crown, S82 OLD, #5540):
//
//	"Reach
//	 Whenever an opponent casts a spell, create a 1/1 green Insect creature
//	 token."
//
// Bloodvial Purveyor's opponent-cast trigger; the token is the Spider's
// controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f17d3aa4-6e88-4973-b2a4-942098a1c463",
		Name:            "Dragonlair Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.B15OpponentCastSpell(ev, source)
				},
				"Dragonlair Spider — create a 1/1 green Insect creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Insect"), N: 1})),
		},
	})
}
