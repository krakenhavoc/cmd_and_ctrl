package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wailing Ghoul — Creature — Zombie {1}{B}, 1/3 (Foundations Jumpstart, S82,
// #5929):
//
//	"When this creature enters, mill two cards."
//
// Aftermath Analyst's mill-on-enter trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ce24851-70d2-416c-8733-78928b4eee74",
		Name:         "Wailing Ghoul",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wailing Ghoul — mill two cards", effects.Do(effects.MillCards{N: 2})),
		},
	})
}
