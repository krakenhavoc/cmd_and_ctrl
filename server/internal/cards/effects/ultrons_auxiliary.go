package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultron's Auxiliary — Artifact Creature — Robot Villain {1}{B}, 1/1 (Marvel
// Super Heroes Commander, #4455):
//
//	"Menace
//	 Whenever another artifact is put into your graveyard from the battlefield
//	 or an artifact card is put into your graveyard from anywhere other than
//	 the battlefield, put a +1/+1 counter on this creature."
//
// Not a "one or more" clause: each artifact is its own occurrence, so two
// artifacts dying together put two counters on it
// (WhenAnotherArtifactIsPutIntoYourGraveyard, graveyard_arrival.go, #4721).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff1ddf25-df25-4733-b9f1-236ad5a08782",
		Name:            "Ultron's Auxiliary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenAnotherArtifactIsPutIntoYourGraveyard("Ultron's Auxiliary — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
