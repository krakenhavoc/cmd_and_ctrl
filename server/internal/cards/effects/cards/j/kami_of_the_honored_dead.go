package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of the Honored Dead — Creature — Spirit {5}{W}{W}, 3/5 (Betrayers of Kamigawa, S82
// OLD b596):
//
//	"Flying
//	 Whenever this creature is dealt damage, you gain that much life.
//	 Soulshift 6"
//
// Metropolis Reformer's lifegain: all the damage dealt at once is one trigger for the total
// (CR 603.2c). Soulshift 6 is the shared keyword trigger (CR 702.46a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "be6d78c7-b7fb-46a9-99ea-f10d21c3da58",
		Name:            "Kami of the Honored Dead",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.B35SelfWasDealtDamage(ev, source)
				},
				Key:        "Kami of the Honored Dead — you gain that much life",
				AtBatchEnd: effects.ThatMuchDamage(),
				Effect:     effects.CmdrB1GainThatMuchDamage,
			},
			effects.Soulshift(6, "Kami of the Honored Dead"),
		},
	})
}
