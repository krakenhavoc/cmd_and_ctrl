package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incite Rebellion — Sorcery {4}{R}{R} (Commander 2014, #5553):
//
//	"For each player, Incite Rebellion deals damage to that player and each
//	 creature that player controls equal to the number of creatures they
//	 control."
//
// Each player's number is counted as the spell resolves, before any damage
// is dealt, and all of it is dealt at once (one damage instance),
// so a creature that dies to it does so at the state-based check afterwards
// and no count shrinks mid-resolution. A player who controls no creature
// takes nothing. The caster is a player too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "26bbc473-cf99-4a82-b432-604757a832c4",
		Name:     "Incite Rebellion",
		// ADR 0126 §6: each creature takes its controller's creature count, so it is declared as the destruction it usually is, an upper bound.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return inciteRebellionDamage(ctx)
		},
	})
}

// inciteRebellionDamage counts each player's creatures, then deals that many
// damage to the player and to each of those creatures, in one instance.
func inciteRebellionDamage(ctx *Context) error {
	creatures := map[uuid.UUID][]uuid.UUID{}
	for _, c := range MatchingBattlefield(ctx, Creature()) {
		creatures[c.Controller] = append(creatures[c.Controller], c.InstanceID)
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, p := range ctx.Game.Seats {
			if p == nil || p.Eliminated {
				continue
			}
			mine := creatures[p.ID]
			for _, id := range append([]uuid.UUID{p.ID}, mine...) {
				if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: len(mine)}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
