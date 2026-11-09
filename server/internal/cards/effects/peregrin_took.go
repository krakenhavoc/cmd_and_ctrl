package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peregrin Took — Legendary Creature — Halfling Citizen {2}{G}, 2/3:
//
//	"If one or more tokens would be created under your control, those
//	 tokens plus an additional Food token are created instead. (It's
//	 an artifact with '{2}, {T}, Sacrifice this token: You gain 3
//	 life.')
//	 Sacrifice three Foods: Draw a card."
//
// The first line is one more CR 701.7b token-creation replacement in
// the family Doubling Season / Parallel Lives / Academy Manufactor
// already live in (#762, ADR 0061) — those multiply or rewrite the
// instruction's groups; this one APPENDS a new one, a single Food,
// to whatever the instruction already makes. It fires once per
// creation instruction, so making three Soldiers plus a Food is one
// application, not three. Because it mutates the SAME event a
// doubler or Academy Manufactor would also see, it composes with
// them exactly as the rules want: a Doubling Season that hasn't
// applied yet still sees (and doubles) the Food this adds.
//
// The second line is `SacrificeN` (Peregrin's own doc comment in
// activated.go names this card) over a Food subtype filter — no new
// vocabulary.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "188da0ad-8524-4fb2-915e-1876dc9df89f",
		Name:         "Peregrin Took",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{peregrinTookFoodReplacement()},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice three Foods: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    SacrificeN(3, "three Foods", HasSubtype("Food")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

// peregrinTookFoodReplacement is "those tokens plus an additional
// Food token are created instead" — an APPEND to the instruction's
// groups, not a multiply or a kind rewrite.
func peregrinTookFoodReplacement() game.ReplacementEffect {
	return plusAnAdditionalTokenReplacement("Peregrin Took: plus an additional Food", FoodToken)
}
