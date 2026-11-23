package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const ygraFoodGrant = "ygra-eater-of-all/food"

// Ygra, Eater of All — Legendary Creature — Cat Elemental {3}{B}{G}, 6/6
// (Bloomburrow, #4467):
//
//	"Ward—Sacrifice a Food.
//	 Other creatures are Food artifacts in addition to their other types
//	 and have '{2}, {T}, Sacrifice this permanent: You gain 3 life.'
//	 Whenever a Food is put into a graveyard from the battlefield, put two
//	 +1/+1 counters on Ygra."
//
// The ward is Ward(WardSacrifice(…)) over a Food artifact. The static is
// two effects of one printed sentence: the layer-4 type add (Artifact and
// the Food subtype, CR 205.1b) and the layer-6 grant of the Food ability
// as a bundle (ADR 0093). The counters trigger watches any Food leaving the
// battlefield for a graveyard, read from its last-known subtypes (CR 603.10a),
// so an opponent's Food creature, a Food token and Ygra's own Food count.
//
// No simplification.
func init() {
	otherCreatures := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.InstanceID != source.InstanceID
	}
	effects.Register(effects.Spec{
		OracleID:        "62511d43-8eaf-405a-9c81-7bd593d3a01e",
		Name:            "Ygra, Eater of All",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: nil,
		Grants: []effects.AbilityGrant{{
			Key: ygraFoodGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}, {T}, Sacrifice this permanent: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{2}, {T}, Sacrifice this permanent: You gain 3 life.",
		}},
		Static: []game.StaticAbility{
			{
				Label:     "Other creatures are Food artifacts in addition to their other types.",
				Layer:     game.Layer4Type,
				AppliesTo: otherCreatures,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Types, "Artifact") {
						c.Types = append(append([]string(nil), c.Types...), "Artifact")
					}
					if !effects.EotHasType(c.Subtypes, "Food") {
						c.Subtypes = append(append([]string(nil), c.Subtypes...), "Food")
					}
				},
			},
			effects.GrantAbilities(otherCreatures, ygraFoodGrant),
		},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardSacrifice("a Food", effects.HasSubtype("Food")), "Ygra, Eater of All — ward, sacrifice a Food"),
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.LeftAsSubtype(ev, c, "Food")
			}, "Ygra, Eater of All — put two +1/+1 counters on Ygra", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
