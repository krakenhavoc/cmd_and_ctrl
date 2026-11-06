package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pirate Peddlers — Creature — Human Pirate {2}{B}, 2/2 (Avatar: The
// Last Airbender, #4459):
//
//	"Deathtouch
//	 Whenever you sacrifice another permanent, put a +1/+1 counter on
//	 this creature."
//
// Juri's "whenever you sacrifice a permanent" (EventSacrifice with you
// as the Actor, one per permanent), narrowed to another permanent: the
// Peddlers' own sacrifice does not trigger it. A sacrifice by another
// player of a permanent you own is theirs, not yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4771a871-576e-4a14-92f0-b5b41c63d615",
		Name:            "Pirate Peddlers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && ev.CardID != uuid.Nil && ev.CardID != source.InstanceID
			}, "Pirate Peddlers — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
