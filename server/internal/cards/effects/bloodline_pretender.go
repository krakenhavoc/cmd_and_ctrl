package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "51d6d164-6e5f-4be3-8f09-433e0b8d4c6b",
		Name:            "Bloodline Pretender",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		AsEnters:        ChooseCreatureTypeAsEnters("Bloodline Pretender"),
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, rprB04AnotherCreatureOfTheChosenTypeEntered,
				"Bloodline Pretender — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
