package effects

// Fear of Isolation — Enchantment Creature — Nightmare {1}{U} 2/3 (#5247):
//
//	"As an additional cost to cast this spell, return a permanent you
//	 control to its owner's hand.
//	 Flying"
//
// The additional cost is the activated abilities' return-to-hand clause on
// the cast's mandatory slot (ReturnPermanentToHandCost): the caster names
// the permanent on cast_spell as return_ids, it goes back to its owner's
// hand with the spell on the stack, and with nothing to return the spell
// can't be cast. Any permanent the caster controls pays, a tapped land
// included (one tapped for this spell's own mana too, CR 601.2g), and one
// they control but don't own goes to its owner's hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0ee28549-5909-4372-8376-34010b7b6654",
		Name:            "Fear of Isolation",
		Completeness:    CompletenessFull,
		AdditionalCost:  ReturnPermanentToHandCost("a permanent you control"),
		PrintedKeywords: []string{"flying"},
	})
}
