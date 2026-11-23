package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tempest Hart // Scan the Clouds — Creature — Elemental Elk {3}{G}, 3/4
// // Instant — Adventure {1}{U} (Wilds of Eldraine, #4471):
//
//	Tempest Hart — "Trample
//	 Whenever you cast a spell with mana value 5 or greater, put a +1/+1
//	 counter on this creature."
//	Scan the Clouds — "Draw two cards, then discard two cards. (Then exile
//	 this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The Hart is Lurking Lizards at five: WheneverYouCast with
// ManaValueGE(5), and the counter lands on the Hart if it is still the
// same permanent. Scan the Clouds is Careful Study's loot (lootOne), the
// discards chosen from the hand after both draws. The Adventure lifecycle
// is the engine's (CR 715.3d).
//
// No simplification.
const TempestHartOracleID = "5005df36-a100-492d-89b9-e221793cfb56"

func init() {
	effects.Register(effects.Spec{
		OracleID:        TempestHartOracleID,
		Name:            "Tempest Hart",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ManaValueGE(5), "Tempest Hart — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     TempestHartOracleID + "#1",
		Name:         "Scan the Clouds",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LootOne(ctx.Game, item, 2)
		},
	})
}
