package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heroes' Bane — Creature — Hydra {3}{G}{G}, 0/0 (Foundations, #4465):
//
//	"This creature enters with four +1/+1 counters on it.
//	 {2}{G}{G}: Put X +1/+1 counters on this creature, where X is its
//	 power."
//
// The entry counters are Kalonian Hydra's self-replacement, so the 0/0
// has them before any state check sees it. X is read as the ability
// resolves, with the layers applied (CR 608.2h): Railway Brawler's
// plusOneCountersEqualToItsPower aimed at the source. A Hydra that has
// left and come back is a new object and gets nothing (CR 400.7), nor
// does one whose power is 0 or less.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bccd722a-b610-4017-93bd-313f08448f5e",
		Name:         "Heroes' Bane",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("+1/+1", 4, "Heroes' Bane: enters with four +1/+1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}{G}: Put X +1/+1 counters on this creature, where X is its power.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{G}{G}"),
			Effect: thisStillHere(func(g *game.Game, item *game.StackItem) error {
				return plusOneCountersEqualToItsPower(g, item, item.SourceCardID)
			}),
		}},
	})
}
