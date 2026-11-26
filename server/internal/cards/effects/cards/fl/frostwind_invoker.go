package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frostwind Invoker — {4}{U} Creature — Merfolk Wizard 3/3 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Flying
//	{8}: Creatures you control gain flying until end of turn.
//
// The creatures are those you control as the ability resolves (CR 611.2c): one that arrives later does not gain flying.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "847587eb-8713-496e-b0c7-41ad6d43b852",
		Name:            "Frostwind Invoker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: Creatures you control gain flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{8}"),
			Effect: effects.Do(effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"flying"},
				Label:    "Frostwind Invoker — creatures you control gain flying",
			}),
		}},
	})
}
