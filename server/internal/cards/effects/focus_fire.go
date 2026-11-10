package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6c4514be-fda7-45e8-83fe-04d49e2b932c",
		Name:         "Focus Fire",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := 2 + countControlled(ctx.Game, ctx.Controller(), focusFireCreatureOrSpacecraft)
			return DealDamage{Source: item.SourceCardID, Target: id, Amount: x}.Apply(ctx)
		},
	})
}

// focusFireCreatureOrSpacecraft is "creatures and/or Spacecraft": a
// permanent that is either, counted once if it is both.
func focusFireCreatureOrSpacecraft(c game.Card) bool {
	return c.IsCreature() || c.HasSubtype("Spacecraft")
}
