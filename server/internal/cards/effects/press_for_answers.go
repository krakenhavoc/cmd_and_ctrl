package effects

import "github.com/google/uuid"

// Press for Answers — Sorcery {1}{U} (Shadows over Innistrad, #5543):
//
//	"Tap target creature. It doesn't untap during its controller's next
//	 untap step.
//	 Investigate. (Create a Clue token. It's an artifact with "{2},
//	 Sacrifice this token: Draw a card.")"
//
// Tap the creature and skip its controller's next untap of it
// (TapAndFreeze, Code of Constraint's addendum), then investigate (CR
// 701.16).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4960c3a0-77ca-48d8-b565-6c5f1e6522cf",
		Name:         "Press for Answers",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Press for Answers — it doesn't untap during its controller's next untap step"}.Apply(ctx)
			}),
			Investigate{N: 1},
		),
	})
}
