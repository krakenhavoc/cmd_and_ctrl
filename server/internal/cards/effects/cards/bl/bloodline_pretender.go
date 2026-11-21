package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodline Pretender — Artifact Creature — Shapeshifter {3}, 2/2:
//
//	"Changeling (This card is every creature type.)
//	 As this creature enters, choose a creature type.
//	 Whenever another creature you control of the chosen type enters, put a
//	 +1/+1 counter on this creature."
//
// The type is chosen as it enters (Metallic Mimic's AsEnters), and the
// trigger reads the chosen type off the source with TribeFilter, which sees
// the entering creature's effective types, so a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "51d6d164-6e5f-4be3-8f09-433e0b8d4c6b",
		Name:            "Bloodline Pretender",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		AsEnters:        effects.ChooseCreatureTypeAsEnters("Bloodline Pretender"),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.RprB04AnotherCreatureOfTheChosenTypeEntered,
				"Bloodline Pretender — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
