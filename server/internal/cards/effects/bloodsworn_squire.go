package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		grant := thisGainsKeywordUntilEndOfTurn("indestructible", name+" — indestructible until end of turn")
		return func(g *game.Game, item *game.StackItem) error {
			if err := grant(g, item); err != nil {
				return err
			}
			return TapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
		}
	}
	Register(Spec{
		OracleID:     oracle,
		Name:         "Bloodsworn Squire",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   label[:len(label)-len("Tap it.")] + "Tap it. Then if there are four or more creature cards in your graveyard, transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}{B}"), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := protect("Bloodsworn Squire")(g, item); err != nil {
					return err
				}
				if b11CreatureCardsInGraveyard(g, item.Controller) < 4 {
					return nil
				}
				return TransformThis{}.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Bloodsworn Knight",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Power and toughness are each equal to the number of creature cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b11CreatureCardsInGraveyard(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}{B}"), DiscardACard()),
			Effect:  protect("Bloodsworn Knight"),
		}},
	})
}
