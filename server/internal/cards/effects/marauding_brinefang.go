package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marauding Brinefang — Creature — Dinosaur {5}{U}{U}, 6/7 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Ward {3}
//	 Islandcycling {2} ({2}, Discard this card: Search your library for an
//	 Island card, reveal it, put it into your hand, then shuffle.)"
//
// Ward {3} is the shared Ward trigger (CR 702.21a). Islandcycling is the
// shared typecycling hand ability (CR 702.29e): an Island card is any
// card with the Island land type, a nonbasic one included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbbd507e-06da-4b47-befc-61db8c92539e",
		Name:         "Marauding Brinefang",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{3}"), "Marauding Brinefang — ward {3}"),
		},
		Activated: []ActivatedAbility{
			Typecycling("Islandcycling", "{2}", "an Island card", IsLandWithSubtype("Island")),
		},
	})
}
