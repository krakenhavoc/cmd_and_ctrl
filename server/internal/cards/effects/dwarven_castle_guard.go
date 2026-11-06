package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dwarven Castle Guard — Creature — Dwarf Soldier {1}{W}, 2/1 (Final
// Fantasy, #4462):
//
//	"When this creature dies, create a 1/1 colorless Hero creature
//	 token."
//
// A leaves-the-battlefield trigger (CR 603.6c, 603.10a) that makes the
// Hero job select makes. It looks back in time, so it fires whatever
// killed the Guard, and the Hero comes for its controller as it died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b51e0ae3-126b-4b7e-9dd1-818a5b8e512a",
		Name:         "Dwarven Castle Guard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Dwarven Castle Guard — create a 1/1 colorless Hero creature token",
				Do(CreateToken{Template: TokenCard(jobSelectHeroToken), N: 1})),
		},
	})
}
