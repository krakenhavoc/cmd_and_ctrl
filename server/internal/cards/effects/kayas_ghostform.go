package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaya's Ghostform — Enchantment — Aura {B}:
//
//	"Enchant creature or planeswalker you control
//	 When enchanted permanent dies or is put into exile, return that
//	 card to the battlefield under your control."
//
// One trigger for two exits. It watches the leaves-the-battlefield
// event and accepts a move to a graveyard (dies) or to exile; a
// bounce, a tuck or a commander going to the command zone is neither.
// The trigger is harvested while the Aura is still attached, because
// the exit is emitted before the state-based action that puts the
// fallen-off Aura into the graveyard.
//
// "Your control" is the Aura's controller's, which is usually the
// card's owner. The card comes back as a new object (CR 400.7), so the
// Aura, which went to the graveyard with its host, is not on it. A
// token host is gone from every zone and returns nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b15c6203-6311-44eb-916d-a4c271960c74",
		Name:         "Kaya's Ghostform",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("enchant creature or planeswalker you control",
			Or(Creature(), Planeswalker()), YouControl()),
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventLTB && source.IsAttachedTo(ev.CardID) &&
					(ev.NewZone == game.ZoneGraveyard || ev.NewZone == game.ZoneExile)
			}, "Kaya's Ghostform — return that card to the battlefield under your control",
				func(g *game.Game, item *game.StackItem) error {
					_, err := g.ReturnToBattlefieldForEffect(item.Trigger.Event.CardID, item.Controller, false)
					if errors.Is(err, game.ErrCardNotFound) {
						return nil
					}
					return err
				}),
		},
	})
}
