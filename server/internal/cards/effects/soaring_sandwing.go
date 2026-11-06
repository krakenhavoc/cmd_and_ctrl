package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soaring Sandwing — Creature — Dinosaur {4}{W}{W}, 3/5 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Flying
//	 When this creature enters, you gain 3 life.
//	 Plainscycling {2} ({2}, Discard this card: Search your library for a
//	 Plains card, reveal it, put it into your hand, then shuffle.)"
//
// Flying is the printed keyword (CR 702.9a). The enters trigger uses the
// stack. Plainscycling is the shared Typecycling constructor (CR
// 702.29e) with "a Plains card": any land with the Plains type, not only
// a basic one, as Angel of the Ruins has it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d840e49-3ca5-4e4f-9152-a365ad2b1773",
		Name:            "Soaring Sandwing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Soaring Sandwing — you gain 3 life", Do(GainLife{Amount: 3})),
		},
		Activated: []ActivatedAbility{
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
		},
	})
}
