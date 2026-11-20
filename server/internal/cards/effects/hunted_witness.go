package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunted Witness — Creature — Human {W}, 1/1 (Guilds of Ravnica, #5525):
//
//	"When this creature dies, create a 1/1 white Soldier creature token with
//	 lifelink."
//
// A dies trigger making the token table's 1/1 white lifelink Soldier
// (Trostani Discordant's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1fde570c-aa96-45ca-a6be-f59c8e4f4b6a",
		Name:         "Hunted Witness",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Hunted Witness — create a 1/1 white Soldier creature token with lifelink",
				Do(CreateToken{Template: TokenCard("1/1 white Soldier with lifelink"), N: 1})),
		},
	})
}
