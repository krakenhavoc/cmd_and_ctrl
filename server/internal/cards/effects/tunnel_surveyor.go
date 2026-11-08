package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tunnel Surveyor — Creature — Human Detective {2}{U}, 2/2 (Duskmourn:
// House of Horror, #4466):
//
//	"When this creature enters, create a 1/1 white Glimmer enchantment
//	 creature token."
//
// An enters trigger making Grand Entryway's Glimmer token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7aae54a6-3011-4867-b5ee-21a2111aecc6",
		Name:         "Tunnel Surveyor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tunnel Surveyor — create a 1/1 white Glimmer enchantment creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Glimmer enchantment"), N: 1})),
		},
	})
}
