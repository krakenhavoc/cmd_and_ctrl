package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Runeflare Trap — Instant — Trap {4}{R}{R} (Zendikar, S82 OLD batch 492,
// #5583):
//
//	"If an opponent drew three or more cards this turn, you may pay {R} rather
//	 than pay this spell's mana cost.
//	 Runeflare Trap deals damage to target player equal to the number of cards
//	 in that player's hand."
//
// The alternative cost is conditional on an opponent's draws this turn (the
// per-turn draw tally, CR 121.2); the damage is read from the target's hand as
// the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "259efd92-416e-42d9-b501-ac1338fd35fc",
		Name:         "Runeflare Trap",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent drew three or more cards this turn, pay {R}",
			ManaCost:  "{R}",
			Condition: anOpponentDrewThreeOrMoreCardsThisTurn,
		}},
		Targets: effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			who := item.Targets[0].ID
			return effects.DealDamage{Source: ctx.Source(), Target: who, Amount: effects.B14HandSize(ctx.Game, who)}.Apply(ctx)
		},
	})
}

func anOpponentDrewThreeOrMoreCardsThisTurn(g *game.Game, controller uuid.UUID) bool {
	for _, p := range g.Seats {
		if p != nil && p.ID != controller && g.TurnTallyFor(p.ID).CardsDrawn >= 3 {
			return true
		}
	}
	return false
}
