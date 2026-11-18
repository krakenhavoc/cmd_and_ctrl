package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duergar Hedge-Mage — Creature — Dwarf Shaman {2}{R/W}, 2/2:
//
//	"When this creature enters, if you control two or more Mountains, you
//	 may destroy target artifact.
//	 When this creature enters, if you control two or more Plains, you may
//	 destroy target enchantment."
//
// Two printed triggers, so two registered ones. Each is an intervening "if"
// (CR 603.4): checked as the creature enters (nothing is put on the stack,
// and no target is chosen, otherwise) and again as the trigger resolves. The
// "may" is answered as it resolves (CR 603.5), after the target was chosen as
// the trigger went on the stack. With both Mountains and Plains, both fire.
//
// No simplification.
func init() {
	hedgeMage := func(subtype, what string, spec *game.TargetSpec) game.TriggeredAbility {
		label := "Duergar Hedge-Mage — if you control two or more " + subtype + "s, you may destroy target " + what
		t := InterveningIf(condControlTwoOrMoreOfSubtype(subtype),
			On(game.EventETB, AllOf(Self, controlTwoOrMoreOfSubtype(subtype)), label, destroyChosenPermanent))
		t.Targets = spec
		return Optional(t, "Duergar Hedge-Mage — destroy target "+what+"?")
	}
	Register(Spec{
		OracleID:     "ecb2fd58-b880-4d9e-93cb-c4d2b28233c9",
		Name:         "Duergar Hedge-Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			hedgeMage("Mountain", "artifact", TargetPermanent("target artifact", Artifact())),
			hedgeMage("Plains", "enchantment", TargetPermanent("target enchantment", Enchantment())),
		},
	})
}
