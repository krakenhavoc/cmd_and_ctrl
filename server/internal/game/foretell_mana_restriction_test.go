package game

import "testing"

// TestForetellManaRestrictions: the foretell special action is paid with the
// foretell spend purpose (CR 702.143a), so mana restricted to "foretell cards
// or cast spells that have foretell" (Niko Defies Destiny) pays the {2} and a
// spell that has foretell, and nothing else (S82 OLD batch 060).
func TestForetellManaRestrictions(t *testing.T) {
	niko := ManaRestrictAnyOf([]string{ManaRestrictForetell}, []string{ManaRestrictSpellForetell})
	foretell := specialActionSpendContext(SpecialActionForetell)
	if !foretell.allows([]string{niko}) {
		t.Error("Niko mana did not pay for the foretell special action")
	}
	if foretell.allows([]string{ManaRestrictCast}) {
		t.Error("cast-only mana paid for the foretell special action")
	}
	if specialActionSpendContext(SpecialActionUnlock).allows([]string{ManaRestrictForetell}) {
		t.Error("foretell mana paid an unlock")
	}
	withForetell := ManaSpendContext{Purpose: SpendPurposeCast, ForetellSpell: true}
	if !withForetell.allows([]string{niko}) {
		t.Error("Niko mana did not pay for a spell with foretell")
	}
	plain := ManaSpendContext{Purpose: SpendPurposeCast}
	if plain.allows([]string{niko}) {
		t.Error("Niko mana paid for a spell without foretell")
	}
	activation := ManaSpendContext{Purpose: SpendPurposeActivate, ForetellSpell: true}
	if activation.allows([]string{niko}) {
		t.Error("Niko mana paid an activation")
	}
}
