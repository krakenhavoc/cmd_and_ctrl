package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stinging Shot — Instant {G}
// (Amonkhet, #5535):
//
//	"Put three -1/-1 counters on target creature with flying.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The -1/-1 counter placer the Commander batch shares, on a
// flying target, plus cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fd7d20d-d44c-443c-a2cf-320f39a3bec4",
		Name:         "Stinging Shot",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with flying", HasKeyword("flying")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return cmdr2B25CountersOnTarget(3)(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
