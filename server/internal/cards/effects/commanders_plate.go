package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Commander's Plate — Artifact — Equipment {1}:
//
//	"Equipped creature gets +3/+3 and has protection from each color
//	 that's not in your commander's color identity.
//	 Equip commander {3}
//	 Equip {5}"
//
// TWO EQUIP ABILITIES, the Blackblade Reforged shape: a cheaper one
// narrowed to the caster's own commander (YourCommander(), new for
// this card — see targets.go), and the plain {5} for any other
// creature. Both are ordinary Spec.Activated entries; nothing new
// there.
//
// SIMPLIFICATION, DECLARED: the protection clause is not implemented.
// "Protection from each color that's not in your commander's color
// identity" needs a static ability that reads the controller's
// commander's colour identity and grants protection from its
// complement — and every commander-identity read in the catalog today
// is the narrower NarrowToCommanderIdentity flag on a mana ability
// (Arcane Signet, Commander's Sphere), which only filters a pipe of
// colour CHOICES. There is no production-facing hook that hands a
// Layer 6 StaticAbility.Apply the controller's commander's colours —
// game.CommanderIdentityForTest exists only for cross-package tests,
// by its own name and doc comment, and reaching for it here would be
// exactly the kind of test-only shortcut that name exists to warn
// off. Until that hook exists, the Plate ships as a plain +3/+3.
//
// The equip abilities and the pump are otherwise the whole card and
// need nothing else.
func init() {
	Register(Spec{
		OracleID:     "cae166de-e681-40a0-83a8-3c17cf40e2fc",
		Name:         "Commander's Plate",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The equipped creature doesn't get protection from colors outside your commander's color identity — only the +3/+3 is here.",
		},
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip commander {3}", "{3}",
				TargetCreature("target creature you control that's your commander", YouControl(), YourCommander())),
			EquipAbility("{5}"),
		},
	})
}
