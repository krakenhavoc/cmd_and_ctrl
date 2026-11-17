package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skulltap — Sorcery {1}{B} (Scourge):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw two cards."
//
// Village Rites at sorcery speed: the creature is sacrificed as the spell
// is cast (CR 601.2h), so it is gone even if the spell is countered, and
// with no creature to sacrifice the spell can't be cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "17c7878f-891f-40a6-80a7-47fcc7b8e29b",
		Name:           "Skulltap",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Draws: 2},
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
