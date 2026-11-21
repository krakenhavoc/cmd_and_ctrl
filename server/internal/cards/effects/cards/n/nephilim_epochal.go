package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nephilim Epochal — Legendary Creature — Elder Nephilim
// {U/R}{R/G}{G/W}{W/B}, 4/4 (Mystery Booster Commander Edition):
//
//	"When Nephilim Epochal enters, search your library for a Nephilim card,
//	 reveal it, put it into your hand, then shuffle.
//	 Nephilim you control get +1/+1 for each other Nephilim you control and
//	 can't be blocked by monocolored creatures."
//
// The bonus is a layer 7c static that counts, for each Nephilim it
// applies to, the OTHER Nephilim its controller controls, so with three
// Nephilim each gets +2/+2. Types are read through the effective view
// (a changeling is a Nephilim). The evasion is a block rule over the
// Nephilim you control against monocolored blockers (exactly one colour;
// a colourless or multicoloured creature may block, CR 105.2).
//
// No simplification.
func init() {
	nephilim := effects.And(effects.Creature(), effects.YouControl(), effects.OfSubtype("Nephilim"))
	effects.Register(effects.Spec{
		OracleID:     "c1198033-47bc-4bee-a3a5-3917dae5cba3",
		Name:         "Nephilim Epochal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nephilim Epochal — search your library for a Nephilim card, reveal it, put it into your hand, then shuffle", func(g *game.Game, item *game.StackItem) error {
				word := effects.NewContext(g, item).TextSubtype("Nephilim")
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype(word) },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Nephilim Epochal — a Nephilim card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Static: []game.StaticAbility{{
			Label:    "Nephilim you control get +1/+1 for each other Nephilim you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Nephilim"))
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
				word := source.TextSubtype("Nephilim")
				n := effects.CountPermanentsYouControl(g, source, func(other *game.Card) bool {
					return other.InstanceID != target.InstanceID && other.IsCreature() && other.HasSubtype(word)
				})
				c.Power += n
				c.Toughness += n
			},
		}},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnMatching(nephilim), effects.And(effects.Creature(), effects.B41Monocolored()), "monocolored creatures"),
		},
	})
}
