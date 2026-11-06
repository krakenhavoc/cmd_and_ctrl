package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran Guardmouse — Creature — Mouse Soldier {3}{R/W}, 3/4 (Bloomburrow,
// #5230):
//
//	"Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, it gets +1/+0 and
//	 gains first strike until end of turn. Scry 1."
//
// The shared valiant trigger (valiant.go). The pump and the first strike
// land only on the permanent that was targeted (a Guardmouse that left and
// came back is a new object, CR 400.7), and the scry happens either way,
// after them, as printed. The scry rider is declared on the card's own
// Purpose, as Attentive Sunscribe's is, since a triggered row's rider is
// not read (#4978).
//
// No simplification.
func init() {
	const label = "Veteran Guardmouse — Valiant: it gets +1/+0 and gains first strike until end of turn; scry 1"
	Register(Spec{
		OracleID:     "2313a3f9-3eac-4684-a25e-6a6ae445a723",
		Name:         "Veteran Guardmouse",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			Valiant(label, Do(
				Instruction(func(ctx *Context) error {
					return socB27ThisGetsAndGains(ctx, 1, "first strike", label)
				}),
				Scry{N: 1},
			)),
		},
	})
}
