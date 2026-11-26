package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrine of Burning Rage — Artifact {2} (New Phyrexia):
//
//	"At the beginning of your upkeep and whenever you cast a red spell,
//	 put a charge counter on this artifact.
//	 {3}, {T}, Sacrifice this artifact: It deals damage equal to the number of charge counters on it to any target."
//
// The counter clause is ONE triggered ability with two trigger conditions.
// The second ability sacrifices the Shrine as a cost, so the number of
// charge counters is its last-known information (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2535383-79f5-4900-8dce-227b0b3650c4",
		Name:         "Shrine of Burning Rage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB460ShrineCounterTrigger("Shrine of Burning Rage", "R"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: It deals damage equal to the number of charge counters on it to any target.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, DamageIsX: true}),
			Effect: effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
				return effects.OldB460ChargeCountersOnSource(ctx.Game, ctx.Source())
			}),
		}},
	})
}
