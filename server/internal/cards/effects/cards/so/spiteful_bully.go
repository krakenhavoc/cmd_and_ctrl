package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spiteful Bully — Creature — Phyrexian Zombie Mercenary {1}{B}, 3/3 (Nemesis,
// S82 OLD, #5618):
//
//	"At the beginning of your upkeep, this creature deals 3 damage to target
//	 creature you control."
//
// The upkeep trigger has a target chosen as it is put on the stack, a creature
// its controller controls, the Bully itself included. The damage is dealt by
// the Bully; a target that is gone or no longer a creature you control takes
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	upkeep := effects.AtYourUpkeep("Spiteful Bully — 3 damage to target creature you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id := effects.FirstLegalCardTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 3}.Apply(ctx)
		})
	upkeep.Targets = effects.TargetCreature("target creature you control", effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     "410acdbc-3a00-49db-a712-ad970d24b363",
		Name:         "Spiteful Bully",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{upkeep},
	})
}
