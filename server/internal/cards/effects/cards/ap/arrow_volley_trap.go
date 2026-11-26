package cardsap

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arrow Volley Trap — Instant — Trap {3}{W}{W} (Duel Decks: Speed vs.
// Cunning, #5555):
//
//	"If four or more creatures are attacking, you may pay {1}{W} rather than
//	 pay this spell's mana cost.
//	 Arrow Volley Trap deals 5 damage divided as you choose among any number
//	 of target attacking creatures."
//
// The trap clause is an alternative cost (CR 118.9) offered while four or
// more creatures, whoever controls them, are attacking. The damage is
// Arc Lightning's divided shape over attacking creatures: each chosen
// target gets at least 1 of the 5 (CR 601.2d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "131dc114-cd48-43d9-93b3-88df85640e09",
		Name:         "Arrow Volley Trap",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("any number of target attacking creatures",
			effects.IsAttackingCreature()).WithCount(0, 0).Dividing(effects.Divide(5)),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If four or more creatures are attacking, pay {1}{W}",
			ManaCost:  "{1}{W}",
			Condition: func(g *game.Game, _ uuid.UUID) bool { return len(effects.AttackingCreatures(g)) >= 4 },
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
