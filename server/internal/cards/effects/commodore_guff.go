package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commodore Guff — Legendary Planeswalker — Guff {1}{U}{R}{W},
// loyalty 5:
//
//	"At the beginning of your end step, put a loyalty counter on another
//	 target planeswalker you control.
//	 +1: Create a 1/1 red Wizard creature token with "{T}: Add {R}.
//	     Spend this mana only to cast a planeswalker spell."
//	 −3: You draw X cards and Commodore Guff deals X damage to each
//	     opponent, where X is the number of planeswalkers you control.
//	 Commodore Guff can be your commander."
//
// The end-step trigger targets as it goes on the stack (CR 603.3d):
// with no other planeswalker you control it is removed, and Guff
// himself is never a legal target (Another).
//
// +1's Wizard is a catalog token (guffWizardToken) whose mana ability
// stamps its {R} with "spend only to cast a planeswalker spell"
// (ManaRestrictCast + ManaRestrictType), Gideon's Memorial's
// restriction: it pays for no ability and no other spell, and the
// auto-tapper leaves it alone.
//
// −3 counts the planeswalkers you control as it resolves, Guff
// included while he is still there (at 5 loyalty he survives a −3), and
// uses the one number for both halves. The damage is dealt by Guff, or
// by his last-known self if he has gone.
//
// "Can be your commander" is deck-building and is read from the oracle
// text by the deck validator, not here.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "7e5a7a1d-dcfb-4ab8-b773-f521cb4d3fa6",
		Name:         "Commodore Guff",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Commodore Guff — put a loyalty counter on another target planeswalker you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id := FirstLegalBattlefieldTarget(ctx)
						if id == uuid.Nil {
							return nil
						}
						return AddCounter{Target: id, Kind: game.CounterLoyalty, N: 1}.Apply(ctx)
					}),
				Another(TargetPermanent("another target planeswalker you control", Planeswalker(), YouControl())),
			),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Create a 1/1 red Wizard creature token with \"{T}: Add {R}. Spend this mana only to cast a planeswalker spell.\"",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: GuffWizardToken(), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "−3: You draw X cards and Commodore Guff deals X damage to each opponent, where X is the number of planeswalkers you control.",
				Cost:   LoyaltyCost(-3),
				Effect: commodoreGuffDrawAndBurn,
			},
		},
	})
}

// commodoreGuffDrawAndBurn is the −3: X is the planeswalkers the
// activator controls now; draw X, then X damage to each opponent.
func commodoreGuffDrawAndBurn(g *game.Game, item *game.StackItem) error {
	x := countControlled(g, item.Controller, func(c game.Card) bool { return c.IsPlaneswalker() })
	if x <= 0 {
		return nil
	}
	if err := (DrawCards{N: x}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	return damageToEachOpponent(g, item, x)
}

// GuffWizardToken is Commodore Guff's 1/1 red Wizard with "{T}: Add {R}.
// Spend this mana only to cast a planeswalker spell."
func GuffWizardToken() game.Card { return tokenFromCatalog(printedGuffWizardToken) }

// printedGuffWizardToken is that Wizard as PRINTED. Its slug is an
// on-disk identity; "wizard" alone is the noncreature-cast pinger.
func printedGuffWizardToken() tokenTemplate {
	return tokenTemplate{
		Slug: "wizard-planeswalker-mana",
		Card: game.Card{
			Name:      "Wizard",
			TypeLine:  "Token Creature — Wizard",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:      true,
			Produced:     "{R}",
			Restrictions: []string{ManaRestrictCast, ManaRestrictType("Planeswalker")},
			Label:        "{T}: Add {R}. Spend this mana only to cast a planeswalker spell",
		}},
		Text: "{T}: Add {R}. Spend this mana only to cast a planeswalker spell.",
	}
}
