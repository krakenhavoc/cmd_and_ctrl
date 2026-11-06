package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battershield Warrior — Creature — Human Warrior {2}{W}, 2/2:
//
//	"Boast — {1}{W}: Creatures you control get +1/+1 until end of turn. (Activate only if
//	 this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves. The row declares its team pump (#4320).
//
// No simplification.
func init() {
	boast := BoastAnswering(game.AnswerPump, "{1}{W}: Creatures you control get +1/+1 until end of turn.",
		ManaCost("{1}{W}"),
		func(g *game.Game, item *game.StackItem) error {
			return BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
				Label: "Battershield Warrior — creatures you control get +1/+1",
			}.Apply(NewContext(g, item))
		})
	boast.Purpose.TeamPump = &game.Pump{Power: 1, Toughness: 1}
	Register(Spec{
		OracleID:     "f8e8cdc6-5685-4986-a14b-20bb9e31d9cf",
		Name:         "Battershield Warrior",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{boast},
	})
}
