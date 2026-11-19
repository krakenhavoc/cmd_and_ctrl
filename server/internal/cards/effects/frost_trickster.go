package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frost Trickster — Creature — Bird Wizard {2}{U}, 2/2:
//
//	"Flying
//	 When this creature enters, tap target creature an opponent controls.
//	 That creature doesn't untap during its controller's next untap step."
//
// Celestial Regulator's tap-and-freeze, unconditional: the ADR 0058 one-shot
// mark rides the tap. A target that is gone or illegal does nothing.
//
// No simplification.
func init() {
	const label = "Frost Trickster — tap target creature an opponent controls; it doesn't untap during its controller's next untap step"
	Register(Spec{
		OracleID:        "1663d54f-4f39-4a03-aac2-e2923cbb7de8",
		Name:            "Frost Trickster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Frost Trickster — doesn't untap during its controller's next untap step"}.Apply(ctx)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
