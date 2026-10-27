package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lunar Avenger — Artifact Creature — Golem {7}, 2/2:
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.)
//	 Remove a +1/+1 counter from this creature: This creature gains
//	 your choice of flying, first strike, or haste until end of turn."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). The keyword is chosen as the ability resolves.
func init() {
	Register(Spec{
		OracleID:            "ab42fb8b-c463-45db-b71d-ed847c9abf1f",
		Name:                "Lunar Avenger",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []ActivatedAbility{{
			Label:   "Remove a +1/+1 counter from this creature: This creature gains your choice of flying, first strike, or haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    RemoveCountersFromThis(game.CounterPlusOne, 1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PickOption{
					Question: "Lunar Avenger gains which until end of turn?",
					Options: []game.ChoiceOption{
						{Label: "Flying"}, {Label: "First strike"}, {Label: "Haste"},
					},
					ThenKey: lunarAvengerAnswer,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// lunarAvengerKeywords are the three choices, in the order the question lists them.
var lunarAvengerKeywords = []string{"flying", "first strike", "haste"}

// lunarAvengerAnswer is the keyword chosen, granted to the Avenger (the
// prompt's source) until end of turn. Keyed (#4253).
var lunarAvengerAnswer = OptionPickThen("option-pick/lunar-avenger-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(lunarAvengerKeywords) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: r.Source, Keywords: []string{lunarAvengerKeywords[r.Index]}}.Apply(ctx)
})
