package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War Screecher — Creature — Bird {1}{W}, 1/3 (Ravnica: Clue Edition):
//
//	"Flying
//	 {5}{W}, {T}: Other creatures you control get +1/+1 until end of turn."
//
// Syr Alin's body: "other creatures you control" is the set as the ability
// resolves (CR 611.2c); War Screecher is never in it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4617739e-4e3e-4d06-b074-8f836261a870",
		Name:            "War Screecher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{W}, {T}: Other creatures you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{5}{W}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)), Power: 1, Toughness: 1,
					Label: "War Screecher — other creatures you control get +1/+1",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
