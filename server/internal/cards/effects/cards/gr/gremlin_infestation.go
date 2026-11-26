package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gremlin Infestation — {3}{R} Enchantment — Aura (Aether Revolt, #5536):
//
//	"Enchant artifact
//	 At the beginning of your end step, this Aura deals 2 damage to enchanted
//	 artifact's controller.
//	 When enchanted artifact is put into a graveyard, create a 2/2 red
//	 Gremlin creature token."
//
// The damage goes to whoever controls the artifact now, read as the ability
// resolves. The second trigger watches the host leave the battlefield for a
// graveyard while this Aura is still attached to it (the Aura's own leaving
// is a state-based action that follows).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94fa21f1-826a-4d51-9ccd-d72e5cf13f3d",
		Name:         "Gremlin Infestation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Artifact()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.StepBegan(game.StepEnd, true),
				"Gremlin Infestation — 2 damage to enchanted artifact's controller",
				func(g *game.Game, item *game.StackItem) error {
					host := effects.AttachedHostFor(g, item.SourceCardID)
					if host == nil {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: host.Controller, Amount: 2}.Apply(effects.NewContext(g, item))
				}),
			effects.WhenEnchantedCreatureDies("Gremlin Infestation — create a 2/2 red Gremlin",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red Gremlin"), N: 1})),
		},
	})
}
