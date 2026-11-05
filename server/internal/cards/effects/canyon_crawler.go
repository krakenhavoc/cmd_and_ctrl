package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Canyon Crawler — Creature — Spider Beast {4}{B}{B}, 6/6:
//
//	"Deathtouch
//	 When this creature enters, create a Food token.
//	 Swampcycling {2} ({2}, Discard this card: Search your library for a
//	 Swamp card, reveal it, put it into your hand, then shuffle.)"
//
// City Pigeon's Food on entry and Bebop's Swampcycling {2}
// (`Typecycling`, CR 702.29e). The Food is the catalog's printed Food
// token.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "bac94fc9-808b-4dc9-b405-408dd2aa6194",
		Name:            "Canyon Crawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Canyon Crawler — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{
			Typecycling("Swampcycling", "{2}", "a Swamp card", IsLandWithSubtype("Swamp")),
		},
	})
}
