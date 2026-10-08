package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rush of Ice — Sorcery {U}:
//
//	"Tap target creature. It doesn't untap during its controller's next
//	 untap step.
//	 Awaken 3—{4}{U}"
//
// ADR 0135 §3 (#2411): Frost Breath's tap-and-freeze on the one creature,
// then the awaken land (CR 702.113a), each target checked on its own (CR
// 608.2b).
//
// No simplifications.
func init() {
	t := TargetCreature("target creature")
	Register(Spec{
		OracleID:     "e9d62416-c4b1-4982-9593-174d46e975c7",
		Name:         "Rush of Ice",
		Completeness: CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			Awaken(3, "{4}{U}", t),
		},
		OnResolve: AwakenAfter(3, func(_ *game.StackItem, ctx *Context) error {
			if id, ok := awakenSpellCardTarget(ctx); ok {
				return TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Rush of Ice"}.Apply(ctx)
			}
			return nil
		}),
	})
}
