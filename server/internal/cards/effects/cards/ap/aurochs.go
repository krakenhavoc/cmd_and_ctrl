package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aurochs — Creature — Aurochs {3}{G}, 2/3 (Ice Age, S82 OLD, #5636):
//
//	"Trample
//	 Whenever this creature attacks, it gets +1/+0 until end of turn for each
//	 other attacking Aurochs."
//
// Rimehorn Aurochs' attack trigger, counting every other attacking Aurochs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3961ef7c-4eb4-482e-9cda-d49d6a29c5a9",
		Name:            "Aurochs",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Aurochs — +1/+0 until end of turn for each other attacking Aurochs",
				effects.OldB560PumpPerOtherAttackingAurochs("Aurochs — +1/+0 for each other attacking Aurochs")),
		},
	})
}
