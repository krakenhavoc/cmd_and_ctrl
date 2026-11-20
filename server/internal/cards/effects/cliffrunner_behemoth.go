package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cliffrunner Behemoth — Creature — Rhino Beast {3}{G}, 5/3 (Conflux,
// #5587):
//
//	"This creature has haste as long as you control a red permanent.
//	 This creature has lifelink as long as you control a white permanent."
//
// Two layer-6 grants (CR 613.1f), each read continuously through
// conB1WhileYouControlA. Haste checked as it is declared as an attacker
// lets it attack the turn it arrives while a red permanent is there.
//
// No simplification.
func init() {
	haste := KeywordGrant(conB1WhileYouControlA("R"), "haste")
	haste.Label = "This creature has haste as long as you control a red permanent."
	lifelink := KeywordGrant(conB1WhileYouControlA("W"), "lifelink")
	lifelink.Label = "This creature has lifelink as long as you control a white permanent."
	Register(Spec{
		OracleID:     "ea0dde32-1638-47f1-bef6-0813794ba35b",
		Name:         "Cliffrunner Behemoth",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{haste, lifelink},
	})
}
