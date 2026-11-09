package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quina, Qu Gourmet — Legendary Creature — Qu {2}{G}, 2/3 (Final Fantasy,
// #4462):
//
//	"If one or more tokens would be created under your control, those
//	 tokens plus a 1/1 green Frog creature token are created instead.
//	 {2}, Sacrifice a Frog: Put a +1/+1 counter on Quina."
//
// Peregrin Took's token-creation replacement (CR 701.7b) with a Frog in
// place of the Food (plusAnAdditionalTokenReplacement): the Frog is appended
// to what the instruction makes, once per instruction, and takes none of the
// instruction's riders (tapped, a later sacrifice). It applies to tokens
// created under Quina's controller's control whoever controls the source of
// the instruction. The sacrifice cost is any Frog the controller controls,
// not only the tokens this makes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2acb15bd-6c89-4ff3-ab32-beac67e9c88a",
		Name:         "Quina, Qu Gourmet",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			plusAnAdditionalTokenReplacement("Quina, Qu Gourmet: plus a 1/1 green Frog", func() game.Card {
				return TokenCard("1/1 green Frog")
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice a Frog: Put a +1/+1 counter on Quina.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet | game.AnswerProtect | game.AnswerSurvive, SourceCounters: 1},
			Cost:    Plus(ManaCost("{2}"), SacrificeN(1, "a Frog", HasSubtype("Frog"))),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
