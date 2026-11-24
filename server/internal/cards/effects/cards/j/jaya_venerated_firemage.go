package cardsj

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jaya, Venerated Firemage — Legendary Planeswalker — Jaya {4}{R}, loyalty 5:
//
//	"If another red source you control would deal damage to a permanent or
//	 player, it deals that much damage plus 1 to that permanent or player
//	 instead.
//	 −2: Jaya deals 2 damage to any target."
//
// The static is a damage replacement (CR 614.1, CR 616.1): the source must be
// red as it deals the damage, controlled by Jaya's controller and not Jaya
// herself. Jaya's own −2 is therefore not boosted. Every damage event with a
// positive amount to a player or a permanent is raised by one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10a87c58-2d8b-4658-b7d6-8dcbb44a2d2f",
		Name:            "Jaya, Venerated Firemage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget == uuid.Nil || ev.DamageSource == src.InstanceID {
					return false
				}
				ch, ok := effects.DamageSourceCharacteristics(ev, g)
				return ok && ch.Controller == src.Controller && slices.Contains(ch.Colors, "R")
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Jaya, Venerated Firemage: +1 damage from another red source you control",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "−2: Jaya deals 2 damage to any target.",
			Cost:    effects.LoyaltyCost(-2),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
			},
		}},
	})
}
