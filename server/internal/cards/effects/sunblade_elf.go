package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunblade Elf — Creature — Elf Warrior {G}, 1/1 (Magic 2015, #5556):
//
//	"This creature gets +1/+1 as long as you control a Plains.
//	 {4}{W}: Creatures you control get +1/+1 until end of turn."
//
// Kird Ape's land-gated pump with a Plains, and Charismatic Vanguard's
// team pump at {4}{W} (s82DmuB01TeamPumpRow): the affected set is fixed
// as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b67e36a-6d74-41f7-b534-f7e3b6bf7017",
		Name:         "Sunblade Elf",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Plains.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Plains")),
			Apply:     pumpBy(1, 1),
		}},
		Activated: []ActivatedAbility{s82DmuB01TeamPumpRow("Sunblade Elf", "{4}{W}")},
	})
}
