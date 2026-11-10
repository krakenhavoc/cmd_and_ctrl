package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jazal Goldmane — Legendary Creature — Cat Warrior {2}{W}{W}, 4/4
// (Foundations, #4465):
//
//	"First strike
//	 {3}{W}{W}: Attacking creatures you control get +X/+X until end of
//	 turn, where X is the number of attacking creatures."
//
// Marshal of the Lost's X (every attacking creature, whoever controls it,
// counted once as the ability resolves, CR 608.2h) on Tyvar, the
// Pummeler's team pump, narrowed to the attacking creatures you control
// then (CR 611.2c). Outside combat X is 0 and nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "609d13dc-c850-4877-8ad3-e84171b58069",
		Name:            "Jazal Goldmane",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}{W}: Attacking creatures you control get +X/+X until end of turn, where X is the number of attacking creatures.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{W}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := len(AttackingCreatures(g))
				if x <= 0 {
					return nil
				}
				return BoostUntilEOT{
					Match:     And(Creature(), ControlledBy(item.Controller), AttackingCreature()),
					Power:     x,
					Toughness: x,
					Label:     "Jazal Goldmane — attacking creatures you control get +X/+X",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
