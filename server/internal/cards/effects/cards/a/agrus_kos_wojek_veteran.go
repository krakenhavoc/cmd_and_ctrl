package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agrus Kos, Wojek Veteran — Legendary Creature — Human Soldier {3}{R}{W}, 3/3
// (Ravnica: City of Guilds, S82 OLD batch 575, #5600):
//
//	"Whenever Agrus Kos attacks, attacking red creatures get +2/+0 and
//	 attacking white creatures get +0/+2 until end of turn."
//
// Blaring Captain's shape: as the trigger resolves, the attacking red
// creatures (any player's) are locked in and get +2/+0, and the attacking white
// creatures are locked in and get +0/+2 (CR 611.2c). A red and white creature
// gets both, and Agrus Kos, red and white and attacking, gets +2/+2.
//
// No simplification.
func init() {
	attackingOf := func(color string) effects.CardPredicate {
		return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
			return c.IsCreature() && c.HasColor(color) && c.AttackingTarget != uuid.Nil
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "e758aecd-6ce1-42f6-9716-32871e3bfa41",
		Name:         "Agrus Kos, Wojek Veteran",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Agrus Kos — attacking red creatures get +2/+0 and attacking white creatures get +0/+2",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Match: attackingOf("R"), Power: 2, Label: "Agrus Kos — +2/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.BoostUntilEOT{Match: attackingOf("W"), Toughness: 2, Label: "Agrus Kos — +0/+2"}.Apply(ctx)
				}),
		},
	})
}
