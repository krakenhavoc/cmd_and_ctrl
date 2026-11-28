package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rock Basilisk — Creature — Basilisk {4}{R}{G}, 4/5 (Mirage, S82 OLD batch 810,
// #5633):
//
//	"Whenever this creature blocks or becomes blocked by a non-Wall creature,
//	 destroy that creature at end of combat."
//
// One trigger for each blocker-attacker pair that names the Basilisk (CR 509.3b,
// 509.3d), fired only when the other creature is not a Wall as it stands. The
// destruction is Lowland Basilisk's delayed trigger at the end of combat step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e7bb687-8243-4377-80b5-67ce6b65f9b7",
		Name:         "Rock Basilisk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB810DestroyBlockPairFoeAtEndOfCombat("Rock Basilisk — destroy that creature at end of combat",
				func(foe game.Card) bool { return !foe.HasSubtypeWord("Wall") }),
		},
	})
}
