package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodsworn Squire // Bloodsworn Knight — {3}{B} Creature — Vampire Soldier
// 3/3 // Creature — Vampire Knight */* (Innistrad: Crimson Vow):
//
//	Front: "{1}{B}, Discard a card: This creature gains indestructible until
//	        end of turn. Tap it. Then if there are four or more creature cards
//	        in your graveyard, transform this creature."
//	Back:  "Bloodsworn Knight's power and toughness are each equal to the
//	        number of creature cards in your graveyard.
//	        {1}{B}, Discard a card: This creature gains indestructible until
//	        end of turn. Tap it."
//
// The discard is a cost, so a creature card discarded to pay counts toward
// the four. The Knight's size is a characteristic-defining ability
// (CR 604.3) in layer 7a.
//
// No simplification.
func init() {
	const oracle = "a6375c5b-ec78-4ee8-b681-8d92797a5b43"
	const label = "{1}{B}, Discard a card: This creature gains indestructible until end of turn. Tap it."
	protect := func(name string) func(g *game.Game, item *game.StackItem) error {
		grant := effects.ThisGainsKeywordUntilEndOfTurn("indestructible", name+" — indestructible until end of turn")
		return func(g *game.Game, item *game.StackItem) error {
			if err := grant(g, item); err != nil {
				return err
			}
			return effects.TapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
		}
	}
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Bloodsworn Squire",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   label[:len(label)-len("Tap it.")] + "Tap it. Then if there are four or more creature cards in your graveyard, transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := protect("Bloodsworn Squire")(g, item); err != nil {
					return err
				}
				if effects.B11CreatureCardsInGraveyard(g, item.Controller) < 4 {
					return nil
				}
				return effects.TransformThis{}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Bloodsworn Knight",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Power and toughness are each equal to the number of creature cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B11CreatureCardsInGraveyard(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.DiscardACard()),
			Effect:  protect("Bloodsworn Knight"),
		}},
	})
}
