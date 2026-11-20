package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stonehorn Chanter — Creature — Rhino Cleric {5}{W}, 4/4 (Magic 2014, #5563):
//
//	"{5}{W}: This creature gains vigilance and lifelink until end of turn. (Attacking doesn't cause it to tap. Damage dealt by it also causes you to gain that much life.)"
//
// Sunmane Pegasus's grant at {5}{W}: both keywords until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22c5187b-e3db-4221-a173-e3b83a581fca",
		Name:         "Stonehorn Chanter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}{W}: This creature gains vigilance and lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{5}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil, "Stonehorn Chanter — vigilance and lifelink",
					game.AddKeywordsMod("vigilance", "lifelink"))
			},
		}},
	})
}
