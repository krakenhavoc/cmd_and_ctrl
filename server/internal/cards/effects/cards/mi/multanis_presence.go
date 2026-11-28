package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Multani's Presence — Enchantment {G}:
//
//	"Whenever a spell you've cast is countered, draw a card."
//
// The countered spell must be one its controller cast, so a copy of a spell
// (never cast) and a countered ability do not count. It triggers once for each
// of your spells that is countered, by anyone.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adad1be7-8d3d-4857-a3b0-be33db4c5746",
		Name:         "Multani's Presence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterSpell, effects.OldB718SpellYouCastWasCountered,
				"Multani's Presence — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
