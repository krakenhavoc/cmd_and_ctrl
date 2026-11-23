package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goddric, Cloaked Reveler — Legendary Creature — Human Noble {1}{R}{R}, 3/3:
//
//	"Haste
//	 Celebration — As long as two or more nonland permanents entered the
//	 battlefield under your control this turn, Goddric is a Dragon with
//	 base power and toughness 4/4, flying, and "{R}: Dragons you control
//	 get +1/+0 until end of turn." (He loses all other creature types.)"
//
// One printed ability, four effects while the celebration count (CR 207.2c)
// is two or more: a layer 4 replacement of his creature types with Dragon
// (supertypes and card types stay), a layer 7b base 4/4, a layer 6 flying
// grant, and a layer 6 grant of the activated ability, so he has it only
// while he is a celebrating Goddric. The pump reaches the Dragons you
// control as it resolves and not those that arrive later (CR 611.2c).
//
// No simplification.
func init() {
	const grantKey = "goddric-cloaked-reveler/dragons-pump"
	celebrating := effects.WhileYouHave(effects.Celebration, effects.SelfCreature)
	const label = "Celebration — As long as two or more nonland permanents entered the battlefield under your control this turn, Goddric is a Dragon with base power and toughness 4/4, flying, and \"{R}: Dragons you control get +1/+0 until end of turn.\""
	typeStatic := game.StaticAbility{
		Label:     label,
		Layer:     game.Layer4Type,
		AppliesTo: celebrating,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			c.SetCreatureTypes(source.TextSubtypes("Dragon"))
		},
	}
	ptStatic := game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7B_Set,
		AppliesTo: celebrating,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power, c.Toughness = 4, 4
		},
	}
	flying := effects.KeywordGrant(celebrating, "flying")
	flying.Label = label
	effects.Register(effects.Spec{
		OracleID:        "35b2b172-5fbc-4e4d-a011-390ee1b6c848",
		Name:            "Goddric, Cloaked Reveler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Activated: []effects.ActivatedAbility{{
				Label:   "{R}: Dragons you control get +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.CmdrB5Dragon()), Power: 1,
						Label: "Goddric — Dragons you control get +1/+0",
					}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{R}: Dragons you control get +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{
			typeStatic, ptStatic, flying,
			effects.Labelled(effects.GrantAbilities(celebrating, grantKey), label),
		},
	})
}
