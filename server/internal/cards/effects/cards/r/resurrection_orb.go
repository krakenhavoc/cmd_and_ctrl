package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Resurrection Orb — Artifact — Equipment {2} (Warhammer 40,000 Commander,
// S82 GAP batch 186, #5948):
//
//	"Equipped creature has lifelink.
//	 Whenever equipped creature dies, return that card to the battlefield
//	 under its owner's control at the beginning of the next end step.
//	 Equip {4}"
//
// The dies trigger schedules a delayed trigger (CR 603.7) that returns the
// card as it was put into the graveyard: a card that has left the graveyard
// by then is a new object and stays where it is (CR 400.7). A token has no
// card to return. The creature was equipped as it died, read from its
// last-known information (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b410f294-cbb6-4016-981f-808fd9b7bbae",
		Name:         "Resurrection Orb",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("lifelink")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{4}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.EquippedCreatureDied(ev, source)
			}, "Resurrection Orb — return that card to the battlefield at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					id := item.Trigger.Event.CardID
					c, ok := g.LookupCardForEffect(id)
					if !ok || c.IsToken() {
						return nil
					}
					return effects.ScheduleDelayedTrigger{
						Label:  "Resurrection Orb — return that card to the battlefield under its owner's control",
						Cards:  []uuid.UUID{id},
						Body:   effects.GapB186ReturnToBattlefieldBody,
						Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
