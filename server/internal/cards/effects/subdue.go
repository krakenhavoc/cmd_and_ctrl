package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Subdue — Instant {G}:
//
//	"Prevent all combat damage that would be dealt by target creature this turn. That creature gets +0/+X until end of turn, where X is its mana value."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the spell resolves (CR 400.7); X is its
// mana value as the spell resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "81bac4b8-277b-415a-9064-a80a68fd7051",
		Name:         "Subdue",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return shieldAgainstTheTargetThen(ctx, true, toughnessByManaValueUntilEOT)
		},
	})
}
