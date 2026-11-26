package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lavaball Trap — Instant — Trap {6}{R}{R} (Zendikar, S82 OLD batch 492,
// #5583):
//
//	"If an opponent had two or more lands enter the battlefield under their
//	 control this turn, you may pay {3}{R}{R} rather than pay this spell's mana
//	 cost.
//	 Destroy two target lands. Lavaball Trap deals 4 damage to each creature."
//
// Whiplash Trap's conditional alternative cost, read from the per-turn count of
// lands that entered under one opponent's control. The two lands are destroyed
// first (each still-legal target), then every creature is dealt 4 damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9cdfb6c-6087-4b0d-b8a5-175f3546b23a",
		Name:         "Lavaball Trap",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent had two or more lands enter the battlefield under their control this turn, pay {3}{R}{R}",
			ManaCost:  "{3}{R}{R}",
			Condition: anOpponentHadTwoOrMoreLandsEnterThisTurn,
		}},
		Targets: effects.TargetPermanent("two target lands", effects.Land()).WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DamageEachMatching(ctx, effects.Creature(), 4)
		},
	})
}

func anOpponentHadTwoOrMoreLandsEnterThisTurn(g *game.Game, controller uuid.UUID) bool {
	for _, p := range g.Seats {
		if p != nil && p.ID != controller && g.EnteredWithCardTypeThisTurn(p.ID, "land") >= 2 {
			return true
		}
	}
	return false
}
