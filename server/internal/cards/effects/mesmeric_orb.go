package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mesmeric Orb — Artifact {2} (EDHREC rank 1082):
//
//	"Whenever a permanent becomes untapped, that permanent's
//	 controller mills a card."
//
// The named blocker on #74, and the card that proves the untap step
// is observable. Two mana, no colour, and it mills the entire table:
// every permanent every player untaps, every untap step, for the
// rest of the game.
//
// "A permanent", unqualified — anyone's, anywhere, for any reason.
// So it watches EventUntapCard rather than anything step-shaped, and
// that is the whole reason the per-permanent event had to be the
// per-permanent event: the untap step is where most of the milling
// comes from, but a Voltaic Key, a Seedborn Muse on someone else's
// turn, or an untap-target effect all say "becomes untapped" too and
// all feed the Orb.
//
// Once per PERMANENT, not once per step — a player untapping eight
// lands mills eight cards, as eight separate triggers. That is what
// the card does, and it is why it is a two-mana mill engine rather
// than a curiosity.
//
// CR 701.26b is a change of state, so a permanent that was already
// untapped does not become untapped and the Orb does not see it;
// the engine's untap primitive enforces that centrally (see
// game/untap.go), which is what keeps the Orb from milling the
// whole board every turn regardless of what was tapped.
//
// "That permanent's controller" is read off the event, captured when
// the untap happened, rather than looked up at resolution: the
// permanent may have left the battlefield in between (it is a
// trigger, and it waits for priority), and the mill still happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03efb4f3-b8e2-4441-824f-886dc40712c4",
		Name:         "Mesmeric Orb",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventUntapCard},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventUntapCard && ev.Actor != uuid.Nil
			},
			Key: "Mesmeric Orb — that permanent's controller mills a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return MillCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
