package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Virtuous Variant — Creature — Spider Human Hero {2}{W}, 1/2 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on target creature you
//	 control."
//
// Hunting Moa's counter body (putPlusOneCounterOnEachLegalTarget) on an
// enters trigger whose target is restricted to your own creatures. It can
// target itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fa4ed027-6e18-4a87-9845-e13f274d9994",
		Name:            "Virtuous Variant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Virtuous Variant — put a +1/+1 counter on target creature you control", effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
