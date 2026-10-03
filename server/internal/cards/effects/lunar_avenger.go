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
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []ActivatedAbility{{
			Label: "Remove a +1/+1 counter from this creature: This creature gains your choice of flying, first strike, or haste until end of turn.",
			Cost:  RemoveCountersFromThis(game.CounterPlusOne, 1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				keywords := []string{"flying", "first strike", "haste"}
				return PickOption{
					Question: "Lunar Avenger gains which until end of turn?",
					Options: []game.ChoiceOption{
						{Label: "Flying"}, {Label: "First strike"}, {Label: "Haste"},
					},
					Then: func(ctx *Context, index int) error {
						if index < 0 || index >= len(keywords) {
							return nil
						}
						return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{keywords[index]}}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
