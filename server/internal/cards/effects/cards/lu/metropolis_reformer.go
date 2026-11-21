package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Metropolis Reformer — Creature — Angel Cleric {2}{W}, 2/3:
//
//	"Flying, vigilance
//	 You have hexproof.
//	 Whenever this creature is dealt damage, you gain that much life."
//
// All damage dealt at once (one combat damage step, one instruction)
// is one trigger for the total (CR 603.2c, ThatMuchDamage).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "baec04a1-3bc9-4618-afab-e42013e36648",
		Name:            "Metropolis Reformer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		PlayerKeywords:  []string{effects.KeywordHexproof},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "Metropolis Reformer — you gain that much life",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     effects.CmdrB1GainThatMuchDamage,
		}},
	})
}
