package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enterprising Scallywag — Creature — Goblin Pirate {1}{R}, 2/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"At the beginning of your end step, if you descended this turn,
//	 create a Treasure token."
//
// "You descended this turn" means a permanent card was put into your
// graveyard from anywhere this turn (CR 700.11), so a token dying does
// not count. The "if" is an intervening if (CR 603.4), checked as your end
// step begins and again as the trigger resolves
// (AtYourEndStepIfDescended). The Treasure is the catalog's token
// (CR 111.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97607bf5-ca7b-4b1b-963e-390faf6e413e",
		Name:         "Enterprising Scallywag",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfDescended("Enterprising Scallywag — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
