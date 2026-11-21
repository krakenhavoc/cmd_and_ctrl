package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cut In — Sorcery {3}{R}:
//
//	"Cut In deals 4 damage to target creature. Create a Young Hero Role
//	 token attached to up to one target creature you control. (If you
//	 control another Role on it, put that one into the graveyard.
//	 Enchanted creature has "Whenever this creature attacks, if its
//	 toughness is 3 or less, put a +1/+1 counter on it.")"
//
// Two target clauses, read by slot (#764): slot 0 takes the damage,
// slot 1 ("up to one") gets the Role. Each is checked on its own at
// resolution (CR 608.2b), so the damage landing on a creature that
// left does not stop the Role, and the reverse. The damage is dealt
// first, as printed, so a creature that dies to it is no longer there
// for the Role.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b205900-ec39-455e-900b-615ece08a07f",
		Name:         "Cut In",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1),
		),
		// Slot 1 is given a Role token, which no amount says.
		Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
		// The Role waits for a prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
					return effects.CreateRoleToken{Role: effects.RoleYoungHero, Host: t.ID}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
