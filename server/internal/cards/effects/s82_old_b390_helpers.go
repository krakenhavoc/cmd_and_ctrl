package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OrdealSacrificeTrigger is the second half of a Theros Ordeal: "When you
// sacrifice this Aura, <effect>". It watches EventSacrifice, which fires
// while the Aura is still on the battlefield, for THIS Aura sacrificed by its
// controller, so it triggers for any sacrifice (the Ordeals' second ruling),
// not only its own: an Aura that falls off as a state-based action is not
// sacrificed and doesn't. A target clause, if the Ordeal has one, is set on
// the returned ability by Targeting.
func OrdealSacrificeTrigger(label string, purpose game.Purpose, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventSacrifice},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && ev.Actor == source.Controller
		},
		Key:     label,
		Purpose: purpose,
		Effect:  effect,
	}
}
