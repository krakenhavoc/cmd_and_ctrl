package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archangel of Tithes — Creature — Angel {1}{W}{W}{W}, 3/5:
//
//	"Flying
//	 As long as this creature is untapped, creatures can't attack you
//	 or planeswalkers you control unless their controller pays {1} for
//	 each of those creatures.
//	 As long as this creature is attacking, creatures can't block
//	 unless their controller pays {1} for each of those creatures."
//
// Both halves are taxes, one per side of combat (#5375).
//
// The attack half is the ADR 0080 attack tax, scoped to "you or
// planeswalkers you control" (ProtectingPlaneswalkers) and gated on
// the Angel being untapped as attackers are declared — read off the
// live source, so an Archangel tapped by the time attacks are declared
// taxes nothing (CR 508.1h prices the declaration as it is made).
//
// The block half is the ADR 0221 block tax: while the Angel is
// attacking, every creature declared as a blocker — against any
// attacker, by any defender — costs {1}, paid as that defender's block
// declaration completes (CR 509.1d). A defender who can't pay takes
// blocks back; a "must block" creature under it is not owed (CR
// 509.1c). Without vigilance the Angel taps to attack, so the two
// halves are rarely live at once, which is the card.
func init() {
	Register(Spec{
		OracleID:        "ff5caed4-0276-476d-8fae-edae2536df7f",
		Name:            "Archangel of Tithes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AttackTaxes: []game.AttackTax{
			archangelOfTithesAttackTax(),
		},
		BlockTaxes: []game.BlockTax{
			BlockTaxWhileAttacking("{1}", "As long as this creature is attacking, creatures can't block unless their controller pays {1} for each of those creatures."),
		},
	})
}

// archangelOfTithesAttackTax is the attack half: {1} per creature
// attacking the Angel's controller or their planeswalkers, while the
// Angel is untapped.
func archangelOfTithesAttackTax() game.AttackTax {
	t := ProtectingPlaneswalkers(AttackTax("{1}", "As long as this creature is untapped, creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures."))
	t.AppliesTo = func(q game.AttackTaxQuery) bool { return !q.Source.Tapped }
	return t
}
