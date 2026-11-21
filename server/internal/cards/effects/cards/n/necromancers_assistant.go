package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necromancer's Assistant — Creature — Zombie {2}{B}, 3/1 (Magic 2015,
// #5556):
//
//	"When this creature enters, mill three cards."
//
// Aftermath Analyst's enters trigger: the controller mills three as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f72f111-450a-4da2-81fd-cbc4f9c1ef9d",
		Name:         "Necromancer's Assistant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Necromancer's Assistant — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
	})
}
