package effects

// Vexing Shusher — Creature — Goblin Shaman {R/G}{R/G}, 2/2:
//
//	"This spell can't be countered.
//	 {R/G}: Target spell can't be countered."
//
// The first line is the printed rider (Spec.CantBeCountered). The
// second is the mark on one spell (ADR 0106 §4 decision 4, #1806): the
// ability targets any spell, anyone's, and as it resolves the spell
// gets a "can't be countered" mark for as long as that object is on
// the stack (CR 400.7). A spell that left in response is skipped
// (CR 608.2b). The mark is on the object, not on a player, so a spell
// that changes controller keeps it; a copy of the spell is a new
// object and does not have it (CR 707.2).
//
// It is an ordinary instant-speed activation with no tap, so it can be
// activated in response to a counterspell, as many times as there is
// mana — every activation after the first does nothing more. The bot
// sees one move per spell on the stack, like any targeted ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a20a7cf8-2075-47ad-9229-36264b112e61",
		Name:            "Vexing Shusher",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Activated: []ActivatedAbility{{
			Label:   "{R/G}: Target spell can't be countered.",
			Cost:    ManaCost("{R/G}"),
			Targets: TargetSpell("target spell"),
			Effect:  Do(MarkTargetSpellsCantBeCountered{From: "Vexing Shusher", Label: "Target spell can't be countered."}),
		}},
	})
}
