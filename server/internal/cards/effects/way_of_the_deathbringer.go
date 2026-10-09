package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Way of the Deathbringer — Legendary Enchantment {2}{B} (Reality
// Fracture, tracker #2795):
//
//	"When Way of the Deathbringer enters, empower Jace 5.
//	 Planeswalkers you control have "[−2]: You may sacrifice a creature.
//	 If you do, create a 4/4 green Beast creature token with trample.""
//
// Empower Jace is the keyword action (ADR 0139). The granted row is an
// ADR 0093 layer-6 bundle with a loyalty cost, handed to every
// planeswalker its controller controls (ADR 0140): the walker pays the
// cost, shares its one loyalty activation a turn with its other rows
// (CR 606.3), and the effect is the activator's.
//
// "You may sacrifice" is a yes/no question and then the sacrifice
// prompt, and the Beast is the sacrifice's continuation, so it only
// appears once a creature has really gone ("if you do").
//
// No simplification.
func init() {
	const grant = "way-of-the-deathbringer/beast"
	Register(Spec{
		OracleID:     "54775625-bf44-4035-969e-7e70578c8b98",
		Name:         "Way of the Deathbringer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Way of the Deathbringer — empower Jace 5", Do(EmpowerJace{N: 5})),
		},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label:   "−2: You may sacrifice a creature. If you do, create a 4/4 green Beast creature token with trample.",
				Cost:    LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  wayOfTheDeathbringerMaySacrifice,
			}},
			Text: "[−2]: You may sacrifice a creature. If you do, create a 4/4 green Beast creature token with trample.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToYourPlaneswalkers(grant)},
	})
}

func wayOfTheDeathbringerMaySacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if countControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() }) == 0 {
		return nil
	}
	return MayChoice{
		Question: "Way of the Deathbringer — sacrifice a creature to create a 4/4 Beast?",
		YesLabel: "Sacrifice a creature",
		NoLabel:  "Don't",
		OnYes: func(ctx *Context) error {
			return ctx.Game.PlayerSacrificesThenForEffect(
				item.SourceCardID, item.Controller,
				sacrificeSpec("a creature", Creature()),
				"Way of the Deathbringer — sacrifice a creature",
				1,
				func(g *game.Game, sacrificed game.PromptedSacrifices) error {
					if sacrificed.Count() == 0 {
						return nil
					}
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("4/4 green Beast with trample"),
						N:          1,
					}.Apply(NewContext(g, item))
				})
		},
	}.Apply(ctx)
}
