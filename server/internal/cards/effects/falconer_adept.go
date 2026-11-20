package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falconer Adept — Creature — Human Soldier {3}{W}, 2/3 (Core Set 2021, #5511):
//
//	"Whenever this creature attacks, create a 1/1 white Bird creature token
//	 with flying that's tapped and attacking."
//
// Hanweir Garrison's attack trigger (whenThisAttacksTokensTappedAndAttacking)
// with one Bird: the token attacks the player the Adept attacked, read off
// the attack event, and was never declared as an attacker (CR 508.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2106eba-15f5-42b5-b797-6582ae94f3de",
		Name:         "Falconer Adept",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			whenThisAttacksTokensTappedAndAttacking("Falconer Adept — create a 1/1 white Bird with flying that's tapped and attacking", "1/1 white Bird with flying", 1),
		},
	})
}
