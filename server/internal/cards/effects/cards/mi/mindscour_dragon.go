package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindscour Dragon — Creature — Dragon {4}{U}{U}, 4/4 (Fate Reforged):
//
//	"Flying
//	 Whenever this creature deals combat damage to an opponent, target player
//	 mills four cards."
//
// The target is chosen as the trigger goes on the stack and may be any player,
// the hit opponent or not; combat damage to a planeswalker or battle does not
// trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c5d326a9-c047-4f77-9f0c-b30daf268f15",
		Name:            "Mindscour Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAnOpponent,
				"Mindscour Dragon — target player mills four cards",
				func(g *game.Game, item *game.StackItem) error { return effects.TargetPlayerMillsEffect(4)(g, item) }),
				effects.TargetPlayer("target player")),
		},
	})
}
