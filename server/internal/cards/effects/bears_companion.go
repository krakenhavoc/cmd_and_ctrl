package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bear's Companion — Creature — Human Warrior {2}{G}{U}{R}, 2/2 (Khans
// of Tarkir, #5554):
//
//	"When this creature enters, create a 4/4 green Bear creature token."
//
// The Earth King's enters trigger, which makes the 4/4 green Bear from
// the token table.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "930248b3-44e6-4c0b-bd7d-4b48323a7272",
		Name:         "Bear's Companion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bear's Companion — create a 4/4 green Bear", Do(CreateToken{Template: TokenCard("4/4 green Bear"), N: 1})),
		},
	})
}
