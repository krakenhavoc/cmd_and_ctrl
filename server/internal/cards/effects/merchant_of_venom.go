package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merchant of Venom — Creature — Cat Warlock {3}{B}, 1/1 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Menace
//	 When this creature enters, each player sacrifices a creature of their
//	 choice.
//	 Whenever a player sacrifices a permanent, put a +1/+1 counter on this
//	 creature."
//
// The edict is the engine's: each player in APNAP order chooses, seeing
// the choices before theirs, and the chosen creatures are sacrificed
// together (the ruling). The Merchant is a creature, so its controller
// may sacrifice it to its own edict. Every sacrifice is its own event, so
// "whenever a player sacrifices a permanent" triggers once for each
// permanent sacrificed, its own controller's included, as Carmen, Cruel
// Skymarcher's does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3918eaa7-5a34-4d6a-be93-a95e733d041b",
		Name:            "Merchant of Venom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Merchant of Venom — each player sacrifices a creature", Do(EachPlayerSacrifices{
				Match: Creature(),
				Label: "a creature",
			}))),
			On(game.EventSacrifice, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID != uuid.Nil
			}, "Merchant of Venom — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
