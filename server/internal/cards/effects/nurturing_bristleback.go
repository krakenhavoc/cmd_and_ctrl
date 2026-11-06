package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nurturing Bristleback — Creature — Dinosaur {5}{G}{G}, 5/5 (The Lost
// Caverns of Ixalan, #4470):
//
//	"When this creature enters, create a 3/3 green Dinosaur creature
//	 token.
//	 Forestcycling {2} ({2}, Discard this card: Search your library for a
//	 Forest card, reveal it, put it into your hand, then shuffle.)"
//
// The enters trigger uses the stack and makes a vanilla 3/3 green
// Dinosaur (the table's row, with no trample). Forestcycling is the shared
// typecycling hand ability (CR 702.29e): a Forest card is any card with
// the Forest land type, a nonbasic one included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d33509d6-9d0c-4248-bb6d-0a9b7712f3a0",
		Name:         "Nurturing Bristleback",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Nurturing Bristleback — create a 3/3 green Dinosaur creature token",
				Do(CreateToken{Template: TokenCard("3/3 green Dinosaur"), N: 1})),
		},
		Activated: []ActivatedAbility{
			Typecycling("Forestcycling", "{2}", "a Forest card", IsLandWithSubtype("Forest")),
		},
	})
}
