package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const crystallineResonanceGrant = "crystalline-resonance/this-ability"

// Crystalline Resonance — Enchantment {2}{U} (Commander 2020, S82 OLD, #5513):
//
//	"Whenever you cycle a card, you may have this enchantment become a copy
//	 of another target permanent until your next turn, except it has this
//	 ability."
//
// Court of Vantress's copy with its duration cut to "until your next turn"
// (CopyUntilYourNextTurn, CR 611.2b) and the exception that keeps this
// trigger: the ability rides a grant bundle added to the copied values
// (CR 707.9a), so the copy cycles on. The target is chosen as the trigger goes
// on the stack, "another" by object identity; the "you may" is asked as it
// resolves.
//
// No simplification.
func init() {
	trigger := effects.Optional(effects.Targeting(
		effects.WheneverYouCycle("Crystalline Resonance — you may have this enchantment become a copy of another target permanent until your next turn",
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.BecomeCopy{
						Targets:  []uuid.UUID{ctx.Source()},
						Of:       t.ID,
						Duration: effects.CopyUntilYourNextTurn,
						Except:   func(v *game.PrintedValues) { v.GrantAbility(crystallineResonanceGrant) },
						Label:    "Crystalline Resonance — becomes a copy until your next turn",
					}.Apply(ctx)
				}
				return nil
			}),
		effects.Another(effects.TargetPermanent("another target permanent"))),
		"Crystalline Resonance — have this enchantment become a copy of the target permanent?")
	effects.Register(effects.Spec{
		OracleID:     "93255c68-2ed1-499b-82c6-672823451ce9",
		Name:         "Crystalline Resonance",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       crystallineResonanceGrant,
			Text:      "Whenever you cycle a card, you may have this enchantment become a copy of another target permanent until your next turn, except it has this ability.",
			Triggered: []game.TriggeredAbility{trigger},
		}},
		Triggered: []game.TriggeredAbility{trigger},
	})
}
