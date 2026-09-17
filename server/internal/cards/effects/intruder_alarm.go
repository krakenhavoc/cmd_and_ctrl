package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intruder Alarm —
//
// "Creatures don't untap during their controllers' untap steps. Whenever a
// creature enters, untap all creatures."
func init() {
	Register(Spec{
		OracleID:              "1e943e04-e213-4781-b1a7-935aad8790e1",
		Name:                  "Intruder Alarm",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringTheirControllersUntapSteps(Creature())},
		Triggered: []game.TriggeredAbility{On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.IsCreature()
		}, "Intruder Alarm — untap all creatures", func(g *game.Game, item *game.StackItem) error {
			for _, id := range b751CreatureIDs(g, uuid.Nil) {
				if err := (UntapTarget{Target: id}).Apply(NewContext(g, item)); err != nil {
					return err
				}
			}
			return nil
		})},
	})
}
