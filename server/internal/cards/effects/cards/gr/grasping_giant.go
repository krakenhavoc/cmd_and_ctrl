package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grasping Giant — Creature — Giant {5}{W}, 5/7 (Theros Beyond Death, #5514):
//
//	"Vigilance
//	 Whenever this creature becomes blocked by a creature, exile that creature until
//	 this creature leaves the battlefield."
//
// One trigger per blocker (EventBlock, CR 509.3c: each creature that blocks it): the
// event's CardID blocks its Target, so the condition is Target == this and the
// creature to exile is the event's CardID. The exile lasts until this leaves the
// battlefield (CR 610.3, ExileUntil{ThisLeaves}); a blocker that has already
// left is not exiled.
//
// No simplification.
func init() {
	const label = "Grasping Giant — exile that creature until this creature leaves the battlefield"
	effects.Register(effects.Spec{
		OracleID:        "d601a28d-a02f-45eb-9dab-a3cf50a87ca4",
		Name:            "Grasping Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBlock && ev.Target == source.InstanceID
			}, label, func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				blocker := item.Trigger.Event.CardID
				if !effects.OnBattlefield(g, blocker) || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.ExileUntil{Target: blocker, ThisLeaves: true, Label: "Grasping Giant — the exiled creature returns when Grasping Giant leaves the battlefield"}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
