package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Specter — Creature — Specter {3}{B}, 2/2 (Mirage, S82 OLD batch 810,
// #5633):
//
//	"Whenever this creature blocks or becomes blocked by a nonblack creature,
//	 destroy that creature at end of combat."
//
// Rock Basilisk's pair trigger with the other filter: the other creature is
// nonblack as it stands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da21c21c-88d8-4a93-b2b4-6d7bb28959c9",
		Name:         "Dread Specter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB810DestroyBlockPairFoeAtEndOfCombat("Dread Specter — destroy that creature at end of combat",
				func(foe game.Card) bool { return !foe.HasColorWord("B") }),
		},
	})
}
