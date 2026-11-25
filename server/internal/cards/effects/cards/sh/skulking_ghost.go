package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skulking Ghost — {1}{B} Creature — Spirit 2/1 (Eternal Masters, S82 OLD,
// #5542):
//
//	"Flying
//	 When this creature becomes the target of a spell or ability, sacrifice it."
//
// Eternal Scourge's watch on EventBecomesTarget, narrowed to the Ghost itself
// and open to a spell or ability of any controller, its own included. The
// trigger goes on the stack as the targets are chosen, so the Ghost is
// sacrificed before the spell or ability that did the targeting resolves; a
// second targeting by the same object only finds it gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cd97f8b4-e751-4b52-987b-a442aef57809",
		Name:            "Skulking Ghost",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self, "Skulking Ghost — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
