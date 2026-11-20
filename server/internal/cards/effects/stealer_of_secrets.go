package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stealer of Secrets — Creature — Human Rogue {2}{U}, 2/2 (Return to
// Ravnica, #5566):
//
//	"Whenever this creature deals combat damage to a player, draw a
//	 card."
//
// Scroll Thief's self-only trigger: the damage's source must be this
// creature. Mandatory, no prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f9535f2-f00c-40d3-9997-d1c0ece15f9f",
		Name:         "Stealer of Secrets",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Stealer of Secrets — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
