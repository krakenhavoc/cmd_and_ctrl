package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gargoyle Sentinel — Artifact Creature — Gargoyle {3}, 3/3 (Magic 2011,
// #5579):
//
//	"Defender
//	 {3}: Until end of turn, this creature loses defender and gains
//	 flying."
//
// Defender is the printed keyword. The ability is one layer-6 effect on
// the Sentinel with its own timestamp (CR 613.1f): it removes defender, as
// Jaded Analyst's trigger does, and grants flying, as Cobalt Golem's row
// does, both ending at cleanup (CR 514.2). A source that left and came back
// is a new object and is not changed (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "893f319f-6e3d-4230-bb65-4f9179b18fad",
		Name:            "Gargoyle Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{3}: Until end of turn, this creature loses defender and gains flying.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil,
					"Gargoyle Sentinel — loses defender and gains flying until end of turn",
					game.RemoveKeywordsMod("defender"), game.AddKeywordsMod("flying"))
			},
		}},
	})
}
