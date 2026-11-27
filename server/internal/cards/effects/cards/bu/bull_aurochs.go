package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bull Aurochs — Creature — Aurochs {1}{G}, 2/1 (Coldsnap, S82 OLD b560):
//
//	"Trample
//	 Whenever this creature attacks, it gets +1/+0 until end of turn for each
//	 other attacking Aurochs."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "edd5b228-f19f-4b27-8b92-75545389da14",
		Name:            "Bull Aurochs",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Bull Aurochs — +1/+0 until end of turn for each other attacking Aurochs",
				effects.OldB560PumpPerOtherAttackingAurochs("Bull Aurochs — +1/+0 for each other attacking Aurochs")),
		},
	})
}
