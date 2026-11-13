package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bereavement — Enchantment {1}{B} (Urza's Saga):
//
//	"Whenever a green creature dies, its controller discards a card."
//
// Any player's creature. "Green" and "its controller" are read from the
// creature as it last existed on the battlefield (CR 603.10a): a creature
// painted green by an effect counts, and a stolen creature's discard is
// its thief's. The word "green" follows a text change on Bereavement.
// The player chooses the card, and a player with an empty hand discards
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fae834ee-2b2e-4f68-b4a4-81592cbecefb",
		Name:         "Bereavement",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftAsColor(ev, dead, source.TextColor("G"))
			}, "Bereavement — its controller discards a card", bereavementDiscard),
		},
	})
}

// bereavementDiscard makes the dead creature's last controller discard a
// card of their choice.
func bereavementDiscard(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	player, known := ev.LeftUnderControlOf()
	if !known {
		if dead, ok := g.LookupCardForEffect(ev.CardID); ok {
			player = dead.Controller
		}
	}
	if player == uuid.Nil {
		return nil
	}
	return discardOne(g, item, player)
}
