package effects

// Cavern of Souls — Land:
//
//	"As this land enters, choose a creature type.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell of the chosen type, and that spell can't be
//	 countered."
//
// The tribal land, and the card that makes the whole of S26's
// per-permanent state necessary: the chosen type is not printed data,
// it is an answer one player gave for one copy of the card, and every
// later activation has to read it back.
//
// Two abilities, not one with a mode. The colorless half is
// unconditional and unrestricted — Cavern is a Wastes that also does
// something else — and the auto-tapper can plan around it freely. The
// coloured half carries the restriction and is therefore invisible to
// the auto-tapper, which is correct: which spell restricted mana is
// for is a decision the planner cannot make.
//
// SANDBOX SIMPLIFICATION — "and that spell can't be countered" is
// INERT, for the reason Path of Ancestry and Delighted Halfling
// already record: nothing in the counter path reads a per-spell
// uncounterable flag, and the mana pool records a token's colour and
// restrictions but not that a particular token paid for a particular
// spell. Both halves of that would have to land together. The
// direction is weaker than printed — the mana works, the protection
// does not — and a player who needs the ruling can decline to counter.
//
// The restriction half is NOT a simplification and is the reason this
// card was worth the engine work: a Cavern named for Elf pays only
// for Elf creature spells, and (CR 702.73a) for changelings, which
// really are Elves.
func init() {
	Register(Spec{
		OracleID: "89ca686a-7c72-4d8f-9290-e89635624a83",
		Name:     "Cavern of Souls",
		AsEnters: ChooseCreatureTypeAsEnters("Cavern of Souls"),
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color (chosen-type creature spells only)",
				// "Any color" flatly — the printed text does not
				// mention the commander's identity, so the pipe keeps
				// all five (NarrowToCommanderIdentity stays off).
				RestrictionsFunc: ChosenTypeManaRestrictions(),
			},
		},
	})
}
