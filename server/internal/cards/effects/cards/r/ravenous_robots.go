package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Robots — {1}{R} Artifact Creature — Robot 2/1 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Whenever you cast an artifact spell, create a 1/1 colorless Robot
//	 artifact creature token.
//	 {R}, {T}: Creature tokens you control gain haste until end of turn."
//
// The cast trigger is Sai, Master Thopterist's (an artifact spell its
// controller casts, read as the spell becomes cast, CR 601.2i), making the
// set's 1/1 Robot. The activation is Ainok Strike Leader's grant over
// creature tokens you control, the set locked as it resolves (CR 611.2c),
// so a token made later this turn does not have haste from it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45873747-f2a4-4d83-9492-bfc97df7b605",
		Name:         "Ravenous Robots",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Artifact(), "Ravenous Robots — create a 1/1 colorless Robot artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Robot artifact"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: Creature tokens you control gain haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Effect: effects.Do(effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.IsTokenPredicate(), effects.YouControl()),
				Keywords: []string{"haste"},
				Label:    "Ravenous Robots — haste",
			}),
		}},
	})
}
