package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shared Discovery — Sorcery {U} (#3868):
//
//	"As an additional cost to cast this spell, tap four untapped
//	 creatures you control.
//	 Draw three cards."
//
// The additional cost is New Blood's (TapUntappedCost): four untapped
// creatures the caster controls, named on cast_spell as teamwork_ids and
// tapped with the spell on the stack. Creatures that entered this turn
// may pay (CR 302.6). Fewer than four untapped creatures and the spell
// cannot be cast (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "dd8e9f28-39da-4336-9259-0e34358749d9",
		Name:           "Shared Discovery",
		Completeness:   CompletenessFull,
		AdditionalCost: TapUntappedCost(4, "four untapped creatures you control", Creature()),
		Purpose:        game.Purpose{Draws: 3},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
