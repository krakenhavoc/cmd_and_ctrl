package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darwin, Adaptive Mutant — Legendary Creature — Mutant Hero {1}{G}, 2/1:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Remove two +1/+1 counters from Darwin: He gains indestructible
//	 until end of turn. (Damage and effects that say "destroy" don't
//	 destroy him.)"
//
// Evolve is the engine's keyword trigger (game/evolve.go, #1805). The
// activated ability is a counter-removal cost (ADR 0020 addendum) and
// an ordinary until-end-of-turn keyword grant; it can be activated any
// time Darwin's controller has priority, as often as there are counters
// to pay with.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4061572b-b7ff-4d06-9f71-9e66e064593d",
		Name:            "Darwin, Adaptive Mutant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Activated: []ActivatedAbility{{
			Label:   "Remove two +1/+1 counters from Darwin: He gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    RemoveCountersFromThis(game.CounterPlusOne, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Darwin, Adaptive Mutant — indestructible until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
