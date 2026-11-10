package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	celebrating := whileYouHave(Celebration, selfCreature)
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
	flying := KeywordGrant(celebrating, "flying")
	flying.Label = label
	Register(Spec{
		OracleID:        "35b2b172-5fbc-4e4d-a011-390ee1b6c848",
		Name:            "Goddric, Cloaked Reveler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Grants: []AbilityGrant{{
			Key: grantKey,
			Activated: []ActivatedAbility{{
				Label:   "{R}: Dragons you control get +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl(), cmdrB5Dragon()), Power: 1,
						Label: "Goddric — Dragons you control get +1/+0",
					}.Apply(NewContext(g, item))
				},
			}},
			Text: "{R}: Dragons you control get +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{
			typeStatic, ptStatic, flying,
			Labelled(GrantAbilities(celebrating, grantKey), label),
		},
	})
}
