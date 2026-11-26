package cardsp

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phytotitan — Creature — Plant Elemental {4}{G}{G}, 7/2 (Magic 2015, #5556):
//
//	"When this creature dies, return it to the battlefield tapped under its
//	 owner's control at the beginning of their next upkeep."
//
// A dies trigger that schedules a delayed trigger (CR 603.7) for the owner's
// next upkeep, naming the card in the graveyard. At that upkeep the card
// comes back tapped under its owner's control if it is still in a graveyard;
// one that was exiled or returned meanwhile is a different object and is left
// alone (CR 400.7). The next upkeep is the owner's own, even when the Titan
// died on their turn after its upkeep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2541e0cd-2cc1-4e47-94e3-030c1cb1a6bd",
		Name:         "Phytotitan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.ThisDied,
				"Phytotitan — return it to the battlefield tapped at the beginning of its owner's next upkeep",
				func(g *game.Game, item *game.StackItem) error {
					owner := uuid.Nil
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
						owner = c.Owner
					}
					if owner == uuid.Nil {
						return nil
					}
					return effects.ScheduleDelayedTrigger{
						At:     game.StepUpkeep,
						Label:  "Phytotitan — return it to the battlefield tapped",
						TurnOf: owner,
						Cards:  []uuid.UUID{item.SourceCardID},
						Body:   effects.OldB356ReturnListedTappedBody,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
