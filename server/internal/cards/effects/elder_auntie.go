package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elder Auntie — Creature — Goblin Warlock {2}{R}, 2/2 (Lorwyn Eclipsed,
// #4458):
//
//	"When this creature enters, create a 1/1 black and red Goblin creature
//	 token."
//
// Sourbread Auntie's token row (1/1 black and red Goblin) without the
// blight: the entry trigger always makes one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "079f0d7f-57a5-4101-a9e3-c5a57148b3fb",
		Name:         "Elder Auntie",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elder Auntie — create a 1/1 black and red Goblin creature token",
				Do(CreateToken{Template: TokenCard("1/1 black and red Goblin"), N: 1})),
		},
	})
}
