package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gravitic Punch — Sorcery {3}{R} (Guilds of Ravnica, #5525):
//
//	"Target creature you control deals damage equal to its power to target
//	 player.
//	 Jump-start"
//
// Bite Down's two-clause shape with a player for the second pick: the creature
// is the damage source (so lifelink, deathtouch and protection see it) and its
// power is read as the spell resolves. Nothing is dealt unless both picks are
// still legal (CR 608.2b). Jump-start is the shared alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90ea4c57-e0ff-4697-912c-c6e2dc873529",
		Name:         "Gravitic Punch",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetPlayer("target player"),
		),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Cmdr2S39JumpStart("{3}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			biter, ok := ctx.ClauseTarget(0)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetPlayer {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok || c.CurrentPower() <= 0 {
				return nil
			}
			return effects.DealDamage{Source: biter.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
		},
	})
}
