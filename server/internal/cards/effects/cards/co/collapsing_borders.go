package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Collapsing Borders — Enchantment {3}{R} (Invasion, #5616):
//
//	"Domain — At the beginning of each player's upkeep, that player gains 1
//	 life for each basic land type among lands they control. Then this
//	 enchantment deals 3 damage to that player."
//
// One trigger on every upkeep. "That player" is the player whose upkeep it is:
// they gain life equal to their own domain, read as the trigger resolves, and
// then take 3 damage from the enchantment. A player at 0 life from the damage
// loses to the state-based actions as usual.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fea3f42-643f-43c8-a9a9-4516dd72af5b",
		Name:         "Collapsing Borders",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Collapsing Borders — that player gains 1 life for each basic land type among lands they control, then Collapsing Borders deals 3 damage to them",
				func(g *game.Game, item *game.StackItem) error {
					p := effects.TriggeringActor(item)
					if p == uuid.Nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.GainLife{Player: p, Amount: effects.DomainCount(g, p)}).Apply(ctx); err != nil {
						return err
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: p, Amount: 3}.Apply(ctx)
				}),
		},
	})
}
