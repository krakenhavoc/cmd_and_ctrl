package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turtle-Duck — Creature — Turtle Bird {G}, 0/4 (Avatar: The Last
// Airbender, #4459):
//
//	"{3}: Until end of turn, this creature has base power 4 and gains
//	 trample."
//
// A layer 7b set of base power only (CR 613.4b), so the toughness stays
// whatever it was, and counters and pumps (7c) still apply on top in
// either order (the rulings); a later set-power effect overwrites it by
// timestamp. Trample is a layer 6 grant. A Turtle-Duck that left and came
// back is a new object the ability does not reach (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e93782e3-1b93-459a-a7de-06c3b6591e6f",
		Name:         "Turtle-Duck",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}: Until end of turn, this creature has base power 4 and gains trample.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil,
					"Turtle-Duck — base power 4 and trample until end of turn",
					game.SetBasePowerMod(4), game.AddKeywordsMod("trample"))
			},
		}},
	})
}
