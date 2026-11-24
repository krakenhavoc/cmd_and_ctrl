package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant's Grasp — Enchantment — Aura {2}{U}{U} (Kaldheim, #5505):
//
//	"Enchant Giant you control
//	 When this Aura enters, gain control of target nonland permanent for
//	 as long as this Aura remains on the battlefield."
//
// The control effect is keyed on the Aura as the object it is now
// (DurationWhileSourceRemains, CR 611.2b): the Aura leaving, or leaving
// and coming back before the trigger resolves, ends or prevents it.
func init() {
	enters := effects.WhenThisEnters("Giant's Grasp — gain control of target nonland permanent for as long as this Aura remains on the battlefield",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			d, ok := effects.DurationWhileSourceRemains(ctx, item.SourceCardID)
			if !ok {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.GainControl{
					Target:     t.ID,
					Controller: item.Controller,
					Duration:   d,
					Label:      "Giant's Grasp — control for as long as this Aura remains",
				}.Apply(ctx)
			}
			return nil
		})
	enters.Targets = effects.TargetPermanent("target nonland permanent", effects.Nonland())
	effects.Register(effects.Spec{
		OracleID:     "bca3f496-ce2a-4715-ae47-5641c33d629e",
		Name:         "Giant's Grasp",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("enchant Giant you control", effects.HasSubtype("Giant"), effects.YouControl()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered:    []game.TriggeredAbility{enters},
	})
}
