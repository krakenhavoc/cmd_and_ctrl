package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hedge Troll — Creature — Troll Cleric {2}{G}, 2/2 (Planar Chaos, #5595):
//
//	"This creature gets +1/+1 as long as you control a Plains.
//	 {W}: Regenerate this creature."
//
// Kird Ape's conditional pump with a land subtype: any permanent you
// control with the Plains subtype counts. Regeneration is on the
// ability's own source (CR 701.19).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "500663c1-390c-4598-b614-4168c4f7debe",
		Name:         "Hedge Troll",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Plains.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Plains")),
			Apply:     pumpBy(1, 1),
		}},
		Activated: []ActivatedAbility{lrwB1RegenerateThisRow("{W}")},
	})
}
