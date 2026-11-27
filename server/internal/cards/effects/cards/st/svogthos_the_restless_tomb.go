package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Svogthos, the Restless Tomb — Land (Duel Decks: Izzet vs. Golgari):
//
//	"{T}: Add {C}.
//	 {3}{B}{G}: Until end of turn, this land becomes a black and green Plant
//	 Zombie creature with "This creature's power and toughness are each equal
//	 to the number of creature cards in your graveyard." It's still a land."
//
// Myth Realized's animate-with-a-granted-static shape: the grant is
// layer 7b (the layer a granted static can occupy, ADR 0093 Decision 10) and
// counts the controller's graveyard live, so it moves as creature cards come
// and go; the land adds the Creature type and the Plant Zombie subtypes and
// sets its colours, and keeps its Land type and its mana ability. A land that
// came under your control this turn can animate but can't attack or tap for
// mana while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	const grantKey = "svogthos/power-toughness-creature-cards"
	const text = "This creature's power and toughness are each equal to the number of creature cards in your graveyard."
	effects.Register(effects.Spec{
		OracleID:     "a34a70b8-02e5-4e8c-a9e7-b21c5a11dddf",
		Name:         "Svogthos, the Restless Tomb",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Static: []game.StaticAbility{{
				Label:    text,
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7B_Set,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := effects.B11CreatureCardsInGraveyard(g, source.Controller)
					c.Power, c.Toughness = n, n
				},
			}},
			Text: text,
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}{G}: Until end of turn, this land becomes a black and green Plant Zombie creature with \"" + text + "\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{3}{B}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				self := item.SourceCardID
				if !effects.OnBattlefield(g, self) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.GrantAbilitiesFor{
					Target: self,
					Keys:   []string{grantKey},
					Also: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Plant", "Zombie")...),
						game.SetColorsMod(ctx.TextColors("B", "G")...),
					},
					Label: "Svogthos, the Restless Tomb — becomes a black and green Plant Zombie creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
