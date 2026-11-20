package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Miscast — Instant {U} (Core Set 2021, #5511):
//
//	"Counter target instant or sorcery spell unless its controller pays {3}."
//
// Flusterstorm's clause (instantOrSorcerySpell) with Complicate's
// counter-unless-paid at {3}: the spell's controller is asked to pay as it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e08c71aa-d35c-4c20-b4db-ae9d36ee3169",
		Name:         "Miscast",
		Completeness: CompletenessFull,
		Targets:      instantOrSorcerySpell("target instant or sorcery spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return counterTheTargetSpellUnlessPaid("Miscast", "{3}")(ctx.Game, item)
		},
	})
}
