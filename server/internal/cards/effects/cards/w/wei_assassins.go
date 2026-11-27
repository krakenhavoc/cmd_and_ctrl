package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wei Assassins — Creature — Human Soldier Assassin {3}{B}{B}, 3/2 (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"When this creature enters, target opponent chooses a creature they control. Destroy that creature."
//
// The opponent is the target; the creature is their choice at resolution
// (S82OldB714WeiAssassinsChoose), so hexproof on the creature does not
// matter and a lone creature is forced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e87dd341-416c-4e93-8eb4-1519ecd2674c",
		Name:         "Wei Assassins",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Wei Assassins — target opponent chooses a creature they control; destroy it", effects.S82OldB714WeiAssassinsChoose),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
