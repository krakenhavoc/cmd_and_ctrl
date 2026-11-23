package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:        "1663d54f-4f39-4a03-aac2-e2923cbb7de8",
		Name:            "Frost Trickster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Frost Trickster — doesn't untap during its controller's next untap step"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
