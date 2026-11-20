package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rushblade Commander — Creature — Azra Warrior {B}{R}, 2/2 (Battlebond,
// #5529):
//
//	"Warriors your team controls have haste."
//
// "Your team" is "you" (CR 102.4): Goblin Warchief's layer-6 keyword grant
// (CR 613.1f) with Warriors, to every Warrior creature the controller
// controls, the Commander included.
//
// No simplification.
func init() {
	haste := TribalKeywordGrant(TribeFilter{Tribes: []string{"Warrior"}, YoursOnly: true}, "haste")
	haste.Label = "Warriors your team controls have haste."
	Register(Spec{
		OracleID:     "14613521-a939-426e-9bd5-183186b51583",
		Name:         "Rushblade Commander",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{haste},
	})
}
