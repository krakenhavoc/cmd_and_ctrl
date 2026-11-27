package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Regeneration — Enchantment — Aura {1}{G} (Eighth Edition, S82 older sets
// batch 617, #5607):
//
//	"Enchant creature
//	 {G}: Regenerate enchanted creature."
//
// Firebreathing's shape: the ability is the Aura's own, so its controller pays
// for it and the shield goes on whatever the Aura is attached to when it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89390a33-b289-4edd-a114-6616e49a49c2",
		Name:         "Regeneration",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				host := effects.EnchantedCreatureOf(ctx)
				if host == uuid.Nil {
					return nil
				}
				return effects.Regenerate{Target: host}.Apply(ctx)
			},
		}},
	})
}
