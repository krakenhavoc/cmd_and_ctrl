package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sultai Flayer — Khans of Tarkir (S82 OLD b348, #5554):
//
//	Creature — Snake Shaman {3}{G}, 3/4: Whenever a creature you control with
//	toughness 4 or greater dies, you gain 4 life.
//
// The dying creature's toughness is its last-known one (CR 603.10a), counters
// and pumps included.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df621822-c811-40fa-99c1-ea9920bcd2e2",
		Name:         "Sultai Flayer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.OldB348YourCreatureWithToughnessFourOrMoreDied,
				"Sultai Flayer — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
