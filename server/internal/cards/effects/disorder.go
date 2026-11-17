package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disorder — Sorcery {1}{R} (Urza's Saga):
//
//	"Disorder deals 2 damage to each white creature and each player who
//	 controls a white creature."
//
// Earthquake's shape with the creatures narrowed to white ones and the
// players to those who control one. Which players control a white
// creature is read once, as the spell resolves and before any damage
// (CR 608.2h), so a player whose only white creature dies to this
// damage is still dealt 2. All of it is one damage event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "586da84b-1384-4da2-983d-340c22b73eb8",
		Name:         "Disorder",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			white := And(Creature(), OfColor("W"))
			var players []uuid.UUID
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				if len(permanentsControlledByMatching(ctx.Game, p.ID, white)) > 0 {
					players = append(players, p.ID)
				}
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := damageEachMatching(ctx, white, 2); err != nil {
					return err
				}
				for _, id := range players {
					if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 2}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
