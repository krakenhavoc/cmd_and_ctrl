package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Axgard Braggart — Creature — Dwarf Warrior {3}{W}, 3/3:
//
//	"Boast — {1}{W}: Untap this creature. Put a +1/+1 counter on it. (Activate only if
//	 this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0d4ab83-8b9b-49cd-86b8-720abd0550f4",
		Name:         "Axgard Braggart",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			BoastAnswering(game.AnswerPump, "{1}{W}: Untap this creature. Put a +1/+1 counter on it.",
				ManaCost("{1}{W}"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (UntapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
						return err
					}
					return plusOneCountersOnThis(1)(g, item)
				}),
		},
	})
}
