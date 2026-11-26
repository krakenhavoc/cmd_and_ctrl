package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Showstopper — Instant {1}{B}{R} (Dragon's Maze, S82 OLD b405, #5564):
//
//	"Until end of turn, creatures you control gain "When this creature dies,
//	 it deals 2 damage to target creature an opponent controls.""
//
// Lightning Volley's grant: the creatures are fixed as the spell resolves
// (CR 611.2c) and each carries the dies trigger for the turn. The trigger
// fires from last-known information (CR 608.2h) and picks its target as it
// goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	const grant = "showstopper/dies-damage"
	effects.Register(effects.Spec{
		OracleID:     "0441a901-31ce-419f-a898-e9a9ad4ab351",
		Name:         "Showstopper",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.TriggerWithPurpose(effects.Targeting(
					effects.WhenThisDies("Showstopper — it deals 2 damage to target creature an opponent controls", effects.SourceDealsDamageToEachLegalTarget(2)),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
					effects.ForTargets(effects.DamageToTarget(0, 2))),
			},
			Text: "When this creature dies, it deals 2 damage to target creature an opponent controls.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match: effects.And(effects.Creature(), effects.YouControl()),
				Keys:  []string{grant},
				Label: "Showstopper — creatures you control gain \"When this creature dies, it deals 2 damage to target creature an opponent controls.\"",
			}.Apply(ctx)
		},
	})
}
