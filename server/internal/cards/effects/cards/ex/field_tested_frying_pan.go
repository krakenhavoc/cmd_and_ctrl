package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Field-Tested Frying Pan — Artifact — Equipment {2}{W} (Tales of
// Middle-earth Commander, #5948):
//
//	"When this Equipment enters, create a Food token, then create a 1/1 white
//	 Halfling creature token and attach this Equipment to it.
//	 Equipped creature has "Whenever you gain life, this creature gets +X/+X
//	 until end of turn, where X is the amount of life you gained."
//	 Equip {2}"
//
// Ancestral Blade's enters trigger with the Food first: the Food is created,
// then the Halfling through the keyed creation whose tail attaches this
// Equipment to it (createTokenAndAttachSource), so a token doubler's order
// prompt never splits the two. The granted trigger is Light of Promise's
// shape: it belongs to the equipped creature, so "you" is its controller and
// "this creature" is the creature, and X is the life gained, read from the
// event as the trigger was queued.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c9e3e6a-0bab-4619-b135-d2dbe824b1f0",
		Name:         "Field-Tested Frying Pan",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Field-Tested Frying Pan — create a Food token, then create a 1/1 white Halfling creature token and attach this Equipment to it",
				effects.Do(
					effects.CreateToken{Template: effects.FoodToken(), N: 1},
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.CreateTokenAndAttachSource(ctx.Game, ctx.Item, "1/1 white Halfling", true)
					}),
				)),
		},
		Grants: []effects.AbilityGrant{{
			Key: fryingPanGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverYouGainLife("Field-Tested Frying Pan — this creature gets +X/+X until end of turn, where X is the amount of life you gained",
					func(g *game.Game, item *game.StackItem) error {
						x := item.Trigger.Event.Amount
						if x <= 0 {
							return nil
						}
						return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: x,
							Label: "Field-Tested Frying Pan — +X/+X"}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever you gain life, this creature gets +X/+X until end of turn, where X is the amount of life you gained.",
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(fryingPanGrant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// fryingPanGrant is the key of the ability the Frying Pan grants.
const fryingPanGrant = "field-tested-frying-pan/gain-life-pump"
