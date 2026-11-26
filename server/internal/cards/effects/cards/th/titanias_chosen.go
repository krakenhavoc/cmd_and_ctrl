package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Titania's Chosen — Creature — Elf Archer {2}{G} 1/1 (Commander 2014):
//
//	"Whenever a player casts a green spell, put a +1/+1 counter on this
//	 creature."
//
// Any player's green spell counts; the spell's colour is read through the
// source's text (a Mind Bend on this card changes the word).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9e72fe7-c4b3-421a-8b2f-bb9448b04a44",
		Name:         "Titania's Chosen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("G")),
				"Titania's Chosen — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
