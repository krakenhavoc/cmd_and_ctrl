package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4960c3a0-77ca-48d8-b565-6c5f1e6522cf",
		Name:         "Press for Answers",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Press for Answers — it doesn't untap during its controller's next untap step"}.Apply(ctx)
			}),
			effects.Investigate{N: 1},
		),
	})
}
