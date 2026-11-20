package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eye Collector — Creature — Faerie {B}, 1/1 (Foundations Jumpstart):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, each player
//	 mills a card."
//
// Flying rides PrintedKeywords. The trigger is Syr Konrad's "each player
// mills a card" (every seat, the controller included, in turn order).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "553c2006-e123-406e-ab4a-1d1f67c7468c",
		Name:            "Eye Collector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Eye Collector — each player mills a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B02EachPlayerMills(g, item, 1) }),
		},
	})
}
