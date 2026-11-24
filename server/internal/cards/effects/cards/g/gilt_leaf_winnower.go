package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gilt-Leaf Winnower — Creature — Elf Warrior {3}{B}{B}, 4/3 (Commander
// Legends):
//
//	"Menace
//	 When this creature enters, you may destroy target non-Elf creature whose
//	 power and toughness aren't equal."
//
// War Priest of Thune's optional targeted enters trigger: the target is chosen
// as the trigger goes on the stack (CR 603.3d), and "you may" is asked as it
// resolves (CR 603.5). The target is a non-Elf creature whose power and
// toughness differ, read as the finished values with power unclamped, so a
// creature at -1/0 is a legal target and a 0/0 is not; it is checked again as
// the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "53b99846-3f80-4e9d-ad3c-52a4906b8823",
		Name:            "Gilt-Leaf Winnower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Gilt-Leaf Winnower — destroy target non-Elf creature whose power and toughness aren't equal", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target non-Elf creature whose power and toughness aren't equal", effects.Not(effects.Subtype("Elf")), effects.PowerAndToughnessDiffer())),
				"Gilt-Leaf Winnower — destroy target non-Elf creature whose power and toughness aren't equal?"),
		},
	})
}
