package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amazing Spider-Girl — Legendary Creature — Spider Human Hero
// {3}{W}{W}, 5/4 (Marvel Super Heroes Commander, #4455):
//
//	"Web-slinging {2}{W} (You may cast this spell for {2}{W} if you also
//	 return a tapped creature you control to its owner's hand.)
//	 Flying, vigilance"
//
// Web-slinging is Spider-Sense's alternative cost: the returned creature
// is a cost, so it is chosen as the spell is cast.
func init() {
	Register(Spec{
		OracleID:        "7a6c4216-7aff-41d4-8d79-bd1ac0b263ae",
		Name:            "Amazing Spider-Girl",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		AlternativeCosts: []game.AlternativeCost{{
			Key:          "web_slinging",
			Label:        "Web-slinging {2}{W}",
			ManaCost:     "{2}{W}",
			ReturnToHand: PermanentYouControl("a tapped creature you control", Creature(), b751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
	})
}
