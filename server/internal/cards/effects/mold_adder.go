package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mold Adder — Creature — Fungus Snake {G}, 1/1 (Foundations, #4465):
//
//	"Whenever an opponent casts a blue or black spell, you may put a
//	 +1/+1 counter on this creature."
//
// Forgotten Ancient's optional cast trigger, narrowed to an opponent's
// spell that is blue, black or both (AnOpponentCast). The "you may" is
// asked as the trigger goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2091eb7e-2357-4ef8-b15e-1e6f82176340",
		Name:         "Mold Adder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, AnOpponentCast(Or(OfColor("U"), OfColor("B"))),
				"Mold Adder — put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
				"Mold Adder — put a +1/+1 counter on it?"),
		},
	})
}
