package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mishra's Bauble — Artifact {0}:
//
//	"{T}, Sacrifice this artifact: Look at the top card of target
//	 player's library. Draw a card at the beginning of the next turn's
//	 upkeep."
//
// Portent's two halves with a one-card look: the delayed draw is
// scheduled for the next upkeep of any turn and is the ability
// controller's (CR 603.7); the look is the controller's, and the card
// goes back on top. If the target is illegal on resolution the whole
// ability does nothing, delayed draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63afc3d1-7653-476e-838f-fc18d4a62a21",
		Name:         "Mishra's Bauble",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Look at the top card of target player's library. Draw a card at the beginning of the next turn's upkeep.",
			Purpose: game.Purpose{Draws: 1},
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var victim uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						victim = t.ID
					}
				}
				if victim == uuid.Nil {
					return nil
				}
				if err := (LookAtLibraryThenPlace{
					Owner:     victim,
					N:         1,
					Placement: game.LibraryPlaceTop,
					Label:     "Mishra's Bauble — put it back on top",
				}).Apply(ctx); err != nil {
					return err
				}
				return ScheduleDelayedTrigger{
					At:    game.StepUpkeep,
					Label: "Mishra's Bauble — draw a card",
					Body:  portentDrawBody,
				}.Apply(ctx)
			},
		}},
	})
}
