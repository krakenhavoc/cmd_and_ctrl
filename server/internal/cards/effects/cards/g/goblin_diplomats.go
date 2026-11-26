package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Diplomats — Creature — Goblin {1}{R}, 2/1 (Magic 2014, S82 OLD batch 401, #5563):
//
//	"{T}: Each creature attacks this turn if able."
//
// Incite War's attack requirement (CR 508.1d) over every creature, every
// player's, read live (game.ScopeCreatures): it changes no characteristic, so
// CR 611.2c does not lock the set and a creature that arrives later in the turn
// has to attack too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84b55e6b-5ce3-49af-95fe-46c029e0260c",
		Name:         "Goblin Diplomats",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Each creature attacks this turn if able.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeCreatures, item.Controller,
					[]game.Mod{game.AddAttackRequirementMod(uuid.Nil)}, effects.DurationUntilEndOfTurn(ctx),
					"Goblin Diplomats — each creature attacks this turn if able")
				return nil
			},
		}},
	})
}
