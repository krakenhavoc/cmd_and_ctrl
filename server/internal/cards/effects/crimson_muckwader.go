package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crimson Muckwader — Creature — Lizard {1}{R}, 2/1 (Magic 2013, #5568):
//
//	"This creature gets +1/+1 as long as you control a Swamp.
//	 {2}{B}: Regenerate this creature. (The next time this creature would
//	 be destroyed this turn, instead tap it, remove it from combat, and
//	 heal all damage on it.)"
//
// Kird Ape's land-type pump and the shared regeneration row (CR
// 701.19a): a shield on the source as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b34f341-42bc-408a-85cc-b2d8797612b8",
		Name:         "Crimson Muckwader",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Swamp.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Swamp")),
			Apply:     pumpBy(1, 1),
		}},
		Activated: []ActivatedAbility{regenerateThisRow("{2}{B}")},
	})
}
