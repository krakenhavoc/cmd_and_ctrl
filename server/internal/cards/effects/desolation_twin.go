package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desolation Twin — 10/10 Creature — Eldrazi for {10} (EDHREC rank
// 4024):
//
//	"When you cast this spell, create a 10/10 colorless Eldrazi
//	 creature token."
//
// Twenty power for ten mana, which is why the big-mana Eldrazi decks
// play it over almost anything else at that price. It is in the batch
// because the trigger is a "when you CAST this spell" ability, and
// the whole point of that wording is what happens when the spell is
// answered: the token is created immediately, from the stack, and a
// Counterspell aimed at the Twin takes only the Twin. The ramp deck
// that gets countered still keeps a 10/10.
//
// # Why it needs FromStack
//
// The trigger harvester scans the battlefield, because that is where
// abilities live (CR 113.6). This ability is not on a permanent — it
// is on a spell, an object that exists only between announce and
// resolution, and by the time the Twin would reach the battlefield
// the cast event is long gone. FromStack is the narrow opt-in that
// lets the harvester see it: EventCast only, and only the card the
// event names. Cascade is the other user.
//
// Consequences the wording buys, all of them free once the trigger
// fires from the stack:
//
//   - The trigger goes on the stack ABOVE the Twin, so it resolves
//     first. The token is on the battlefield before the Twin is.
//   - Countering the Twin does not counter the trigger; they are
//     separate objects on the stack.
//   - A Twin that is somehow put onto the battlefield without being
//     cast (reanimated, Elvish Piper'd) makes no token at all. "Cast"
//     means cast.
//
// The controller is the CASTER (the cast event's actor), not the
// card's Controller field, because at that moment the Twin is a spell
// its caster controls — the same reason cascade reads ev.Actor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0cb1f37-1679-42c7-a794-81e088157eeb",
		Name:         "Desolation Twin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			// A fill-in Build (ADR 0041 P9): the controller is the CASTER
			// (ev.Actor), a fact of the moment the spell was cast, not the
			// spell card's own Controller field.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Desolation Twin — create a 10/10 colorless Eldrazi", nil)
				item.Controller, item.Owner = ev.Actor, ev.Actor
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("10/10 colorless Eldrazi"),
					N:          1,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
