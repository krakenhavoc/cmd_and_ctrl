package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lavalanche — Sorcery {X}{B}{R}{G}:
//
//	"Lavalanche deals X damage to target player or planeswalker and each
//	 creature that player or that planeswalker's controller controls."
//
// X rides the cast (ctx.X()). The player is the target, or the controller of
// the targeted planeswalker, read as the spell resolves and before any damage
// lands (a planeswalker the damage kills is no longer there to say who
// controlled it). The creatures are snapshotted first and every point is one
// simultaneous damage instance (CR 608.2). With its one target illegal the
// spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4475e9a7-4e94-4965-8e7e-211d40a98217",
		Name:         "Lavalanche",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      targetPlayerOrPlaneswalker(),
		Purpose: game.Purpose{
			Targets: game.ForTargets(DamageXToTarget(0)),
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, OpponentsOnly: true, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			targets := ctx.LegalTargets()
			if x <= 0 || len(targets) == 0 {
				return nil
			}
			t := targets[0]
			player := t.ID
			if t.Kind == game.TargetCard {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				player = c.Controller
			}
			var creatures []uuid.UUID
			for _, c := range MatchingBattlefield(ctx, Creature()) {
				if c.Controller == player {
					creatures = append(creatures, c.InstanceID)
				}
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}).Apply(ctx); err != nil {
					return err
				}
				for _, id := range creatures {
					if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: x}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
