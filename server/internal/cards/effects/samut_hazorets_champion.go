package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Samut, Hazoret's Champion — Legendary Creature — Human Warrior Cleric
// {1}{R}, 2/2:
//
//	"Creatures you control have haste."
//
// Fervor's grant on a body: every creature the controller has, Samut
// included, from the moment it arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "add02a36-8811-438b-a8e1-f5c8a8146b4e",
		Name:         "Samut, Hazoret's Champion",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, "haste"),
		},
	})
}
