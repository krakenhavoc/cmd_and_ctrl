package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mouser Foundry — {1}{R} Artifact (Teenage Mutant Ninja Turtles, #4457):
//
//	"When this artifact enters or leaves the battlefield, create a 1/1
//	 colorless Robot artifact creature token.
//	 {4}{R}, Sacrifice this artifact: It deals 3 damage to target
//	 creature."
//
// The trigger is Subterranean Shambler's enters-or-leaves ability; leaving
// is any exit and looks back in time (CR 603.10a), so sacrificing the
// Foundry to its own ability makes a second Robot. The damage is Goblin
// Cratermaker's sacrificed source dealing damage on its last-known
// information (CR 113.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "404e958b-ead9-4ec7-b786-7b3d65e29967",
		Name:         "Mouser Foundry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrLeaves("Mouser Foundry — create a 1/1 colorless Robot artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Robot artifact"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}, Sacrifice this artifact: It deals 3 damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{4}{R}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
