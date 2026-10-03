package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shieldmage Elder — Creature — Human Cleric Wizard {5}{W}:
//
//	"Tap two untapped Clerics you control: Prevent all damage target creature would deal this turn.
//	 Tap two untapped Wizards you control: Prevent all damage target spell would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): each shield's source is its
// target, pinned as the ability resolves (CR 400.7). A targeted permanent
// spell's shield also covers the permanent it becomes (its ruling).
//
// The costs tap two Clerics (or Wizards) the activator controls, the
// Elder among them if they like: it is both and may pay its own costs, and tapping for
// them is not the {T} symbol, so creatures that arrived this turn may pay
// (its ruling; CR 302.6).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "943bbe5b-76cd-478d-a1fe-d4e557df3469",
		Name:         "Shieldmage Elder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			shieldAgainstTargetsRow(
				"Tap two untapped Clerics you control: Prevent all damage target creature would deal this turn.",
				tapTwoUntapped("two untapped Clerics you control", "Cleric"), TargetCreature("target creature"), false),
			shieldAgainstTargetsRow(
				"Tap two untapped Wizards you control: Prevent all damage target spell would deal this turn.",
				tapTwoUntapped("two untapped Wizards you control", "Wizard"), TargetSpell("target spell"), false),
		},
	})
}

// tapTwoUntapped is "Tap two untapped <creature type>s you control" as a
// cost, the source included when it has the type.
func tapTwoUntapped(label, creatureType string) game.AbilityCost {
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:  2,
		Filter: TargetPermanent(label, Creature(), OfCreatureType(creatureType)),
		Label:  label,
	}}
}
