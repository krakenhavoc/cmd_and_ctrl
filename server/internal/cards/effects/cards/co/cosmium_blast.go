package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cosmium Blast — Instant {1}{W} (The Lost Caverns of Ixalan, #4470):
//
//	"Cosmium Blast deals 4 damage to target attacking or blocking
//	 creature."
//
// The target clause is Elspeth's Smite's AttackingOrBlocking, which reads
// the combat designations, so the spell has a legal target only during a
// combat with an attacker or a blocker. The damage source is the spell itself
// (CR 120.3e); a target no longer legal, such as one removed from
// combat in response, is dealt nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52c96b5a-ce47-4bde-b530-0033b9bfd808",
		Name:         "Cosmium Blast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 4}.Apply(ctx)
		},
	})
}
