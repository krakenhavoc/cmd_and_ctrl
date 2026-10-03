package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_counter_unless.go — "Whenever a player casts a <kind of> spell,
// counter it unless that player pays <cost>" (Nether Void's {3}, In the
// Eye of Chaos's "{X}, where X is its mana value").
//
// The trigger is the source permanent's and watches every player's
// casts, the source's controller's included ("a player"). It resolves
// above the spell, so the spell is still on the stack and the caster is
// asked the CR 118.12 question through CounterUnlessPaid, which stops
// the table while it is open (#951). The spell's stack ID is its
// card's instance ID, which is what EventCast carries; a spell that
// has already left the stack raises no prompt.
//
// Append-only: add a helper below the last one.

// WheneverAPlayerCastsCounterUnlessPays is the trigger. `match` narrows
// the spell (nil for any spell), and `cost` prices the payment as the
// trigger resolves from the spell as it is on the stack.
func WheneverAPlayerCastsCounterUnlessPays(label string, match func(g *game.Game, spell game.Card) bool,
	cost func(g *game.Game, spell game.Card) string,
) game.TriggeredAbility {
	return On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.CardID == uuid.Nil {
			return false
		}
		if match == nil {
			return true
		}
		spell, ok := g.LookupCardForEffect(ev.CardID)
		return ok && match(g, spell)
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ev := ctx.Trigger().Event
		spell, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || g.StackItemForEffect(ev.CardID) == nil {
			return nil
		}
		return CounterUnlessPaid{StackID: ev.CardID, Cost: cost(g, spell), Question: label}.Apply(ctx)
	})
}

// payTheSpellsManaValue is "{X}, where X is its mana value" — the
// spell's mana value on the stack, {X} included (CR 202.3e).
func payTheSpellsManaValue(g *game.Game, spell game.Card) string {
	mv, ok := g.ManaValueForEffect(spell)
	if !ok || mv < 0 {
		mv = 0
	}
	return "{" + strconv.Itoa(mv) + "}"
}
