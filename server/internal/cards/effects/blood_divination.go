package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Divination — Sorcery {3}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw three cards."
//
// Altar's Reap at sorcery speed and one card more. The sacrifice is paid
// while casting (CR 601.2h), so the creature is gone even if the spell
// is countered, and the cast is refused with no creature to sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "c1cf9d74-e700-456a-9e5a-1c4df22db268",
		Name:           "Blood Divination",
		Purpose:        game.Purpose{Draws: 3, Answers: game.AnswerSacOutlet},
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
