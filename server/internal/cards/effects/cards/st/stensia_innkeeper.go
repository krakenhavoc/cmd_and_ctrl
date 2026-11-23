package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stensia Innkeeper — Creature — Vampire {3}{R}, 3/3 (Eldritch Moon,
// #5541):
//
//	"When this creature enters, tap target land an opponent controls. That
//	 land doesn't untap during its controller's next untap step."
//
// Frost Trickster's tap-and-freeze aimed at a land (b751TapFreezeTargets,
// the land skipping its own controller's next untap step).
//
// No simplification.
func init() {
	const label = "Stensia Innkeeper — tap target land an opponent controls; it doesn't untap during its controller's next untap step"
	effects.Register(effects.Spec{
		OracleID:     "8a7f446b-557a-41cc-b9ef-369e220a0eeb",
		Name:         "Stensia Innkeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
					return effects.B751TapFreezeTargets(effects.NewContext(g, item), uuid.Nil, "Stensia Innkeeper")
				}),
				effects.TargetPermanent("target land an opponent controls", effects.Land(), effects.OpponentControls())),
		},
	})
}
