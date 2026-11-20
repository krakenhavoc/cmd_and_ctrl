package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spire Patrol — Creature — Human Soldier {2}{W}{U}, 3/2 (Aether Revolt,
// #5536):
//
//	"Flying
//	 When this creature enters, tap target creature an opponent controls.
//	 That creature doesn't untap during its controller's next untap step."
//
// Frost Trickster's tap-and-freeze enters trigger (TapAndFreeze, the
// ADR 0058 one-shot mark).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "afacd2b1-1733-4b28-b834-71a31149fa00",
		Name:            "Spire Patrol",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Spire Patrol — tap target creature an opponent controls; it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						return b751TapFreezeTargets(NewContext(g, item), uuid.Nil, "Spire Patrol — doesn't untap during its controller's next untap step")
					}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
