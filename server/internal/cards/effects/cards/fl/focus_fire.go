package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Focus Fire — Instant {W} (Edge of Eternities, #4461):
//
//	"Focus Fire deals X damage to target attacking or blocking creature,
//	 where X is 2 plus the number of creatures and/or Spacecraft you
//	 control."
//
// Cosmium Blast's target clause (AttackingOrBlocking) with Thunder
// Salvo's "2 plus" amount. X is counted as the spell resolves
// (CR 608.2h): each permanent you control that is a creature, a
// Spacecraft or both counts once, so a stationed Spacecraft creature is
// not counted twice and one below its threshold still counts. The damage
// source is the spell (CR 120.3e); a target no longer legal, such as one
// removed from combat in response, is dealt nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c4514be-fda7-45e8-83fe-04d49e2b932c",
		Name:         "Focus Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := 2 + effects.CountControlled(ctx.Game, ctx.Controller(), focusFireCreatureOrSpacecraft)
			return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: x}.Apply(ctx)
		},
	})
}

// focusFireCreatureOrSpacecraft is "creatures and/or Spacecraft": a
// permanent that is either, counted once if it is both.
func focusFireCreatureOrSpacecraft(c game.Card) bool {
	return c.IsCreature() || c.HasSubtype("Spacecraft")
}
