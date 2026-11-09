package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goddric, Cloaked Reveler — Legendary Creature — Human Noble {1}{R}{R},
// 3/3 (Wilds of Eldraine, #4471):
//
//	"Haste
//	 Celebration — As long as two or more nonland permanents entered the
//	 battlefield under your control this turn, Goddric is a Dragon with
//	 base power and toughness 4/4, flying, and "{R}: Dragons you control
//	 get +1/+0 until end of turn." (He loses all other creature types.)"
//
// One printed sentence, four effects, each in its own layer while
// celebration holds (CR 207.2c, celebration.go): layer 4 replaces his
// creature types with Dragon (CR 205.1a, so Human and Noble go), layer 6
// gives him flying and the activated ability as a granted bundle
// (ADR 0093), layer 7b sets his base power and toughness to 4/4 (counters
// and pumps still add). The 7b half is flagged ContinuesAfterRemoval: if
// he loses all abilities in the same turn he is still a Dragon 4/4 (the
// first ruling, CR 613.6), though he loses the flying and the ability.
// The ability's pump is locked at resolution on the Dragons you control
// then (the second ruling), himself included.
//
// No simplification.
func init() {
	const grant = "goddric-cloaked-reveler/dragons-pump"
	celebrating := whileYouHave(Celebration, selfCreature)
	Register(Spec{
		OracleID:        "35b2b172-5fbc-4e4d-a011-390ee1b6c848",
		Name:            "Goddric, Cloaked Reveler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label:   "{R}: Dragons you control get +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{R}"),
				Effect: Do(Instruction(func(ctx *Context) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl(), HasSubtype("Dragon")),
						Power: 1,
						Label: "Goddric, Cloaked Reveler — Dragons get +1/+0",
					}.Apply(ctx)
				})),
			}},
			Text: "{R}: Dragons you control get +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{
			{
				Label:     "Celebration — Goddric is a Dragon (he loses all other creature types).",
				Layer:     game.Layer4Type,
				AppliesTo: celebrating,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.SetSubtypes([]string{"Dragon"})
				},
			},
			SelfKeywordWhileCelebration("flying"),
			GrantAbilities(celebrating, grant),
			{
				Label:                 "Celebration — Goddric has base power and toughness 4/4.",
				Layer:                 game.Layer7PT,
				SubLayer:              game.SubLayer7B_Set,
				ContinuesAfterRemoval: true,
				AppliesTo:             celebrating,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 4, 4
				},
			},
		},
	})
}
