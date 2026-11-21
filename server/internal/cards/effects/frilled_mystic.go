package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frilled Mystic — Creature — Elf Lizard Wizard {G}{G}{U}{U}, 3/2
// (Ravnica Remastered, #5948):
//
//	"Flash
//	 When this creature enters, you may counter target spell."
//
// Flash is the engine keyword (CR 702.8). The enters trigger is Complicate's
// cycle trigger with an unconditional counter: its target is chosen as it goes
// on the stack, so with no spell to target it is removed (CR 603.3d), and the
// "you may" is asked as it resolves (CR 603.5). A spell that left the stack or
// can't be countered is left alone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "724bc69c-cd09-4ff8-91e7-70500b9474e1",
		Name:            "Frilled Mystic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(WhenThisEnters("Frilled Mystic — you may counter target spell", counterTheChosenSpell),
					"Frilled Mystic — counter target spell?"),
				TargetSpell("target spell"),
			),
		},
	})
}
