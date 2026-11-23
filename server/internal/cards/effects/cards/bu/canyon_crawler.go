package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "bac94fc9-808b-4dc9-b405-408dd2aa6194",
		Name:            "Canyon Crawler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Canyon Crawler — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Swampcycling", "{2}", "a Swamp card", effects.IsLandWithSubtype("Swamp")),
		},
	})
}
