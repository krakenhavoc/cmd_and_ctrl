package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Asmoranomardicadaistinaculdacar — Legendary Creature — Human Wizard, no
// mana cost, 3/3 (Modern Horizons 2, #5501):
//
//	"As long as you've discarded a card this turn, you may pay {B/R} to
//	 cast this spell.
//	 When Asmoranomardicadaistinaculdacar enters, you may search your
//	 library for a card named The Underworld Cookbook, reveal it, put it
//	 into your hand, then shuffle.
//	 Sacrifice two Foods: Target creature deals 6 damage to itself."
//
// The card has no mana cost, so it cannot be cast at all except through
// the alternative cost (CR 118.9): {B/R}, offered only while its caster has
// discarded a card this turn (the turn's discard tally, CR 701.9a). The
// enters trigger is a "may" tutor for the one named card, shown to the
// table. The activated ability sacrifices two Foods as a cost and makes the
// target creature the source of damage to itself, six, so its own
// deathtouch or lifelink applies and a protection from it does not matter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f74e26f-5cf3-4b0e-bcc6-f858599474fc",
		Name:         "Asmoranomardicadaistinaculdacar",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:      "asmoranomardicadaistinaculdacar-discarded",
			Label:    "If you've discarded a card this turn, pay {B/R}",
			ManaCost: "{B/R}",
			Condition: func(g *game.Game, controller uuid.UUID) bool {
				return g.TurnTallyFor(controller).CardsDiscarded > 0
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.WhenThisEnters("Asmoranomardicadaistinaculdacar — search your library for a card named The Underworld Cookbook, reveal it, put it into your hand, then shuffle",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.SearchLibrary{
							Player:    ctx.Controller(),
							Predicate: func(c game.Card) bool { return c.Name == "The Underworld Cookbook" },
							Dest:      game.ZoneHand,
							Limit:     1,
							Reveal:    true,
							Shuffle:   true,
							Reason:    "Asmoranomardicadaistinaculdacar — a card named The Underworld Cookbook",
						}.Apply(ctx)
					}),
				"Asmoranomardicadaistinaculdacar — search your library for The Underworld Cookbook?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice two Foods: Target creature deals 6 damage to itself.",
			Cost:    effects.SacrificeN(2, "two Foods", effects.HasSubtype("Food")),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.DealDamage{Source: id, Target: id, Amount: 6}.Apply(ctx)
			},
		}},
	})
}
