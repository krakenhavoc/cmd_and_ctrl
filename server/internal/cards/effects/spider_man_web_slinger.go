package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Man, Web-Slinger — Legendary Creature — Spider Human Hero {2}{W},
// 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Web-slinging {W} (You may cast this spell for {W} if you also return
//	 a tapped creature you control to its owner's hand.)"
//
// Spider-Sense's web-slinging alternative cost: {W} paid in mana and a
// tapped creature you control returned to its owner's hand as a cost, at
// announce. Cast for {2}{W} it needs neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "505231d8-714f-4c5b-a635-d1662bd18a78",
		Name:         "Spider-Man, Web-Slinger",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:          "web_slinging",
			Label:        "Web-slinging {W}",
			ManaCost:     "{W}",
			ReturnToHand: PermanentYouControl("a tapped creature you control", Creature(), b751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
	})
}
