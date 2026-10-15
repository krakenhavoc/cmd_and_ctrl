package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// The restriction half is the reason this card was worth the S26
// engine work: a Cavern named for Elf pays only for Elf creature
// spells, and (CR 702.73a) for changelings, which really are Elves.
//
// "And that spell can't be countered" is a SPEND RIDER (#1547, ADR
// 0040's 2026-09-24 amendment): it rides on the token, and when the
// token pays for a spell the spell's stack item is marked uncounterable
// — read by the one gate every counter verb asks. The rider's filter
// is just "a spell": the restriction has already made sure the only
// spell this mana can pay for is a creature of the chosen type. A
// spell paid for with the COLOURLESS half, or with other mana, is
// counterable as normal.
//
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, so this
// mana pays whenever it can and its rider fires. Mana that never
// entered the pool was not this card's mana.
func init() {
	Register(Spec{
		OracleID:       "89ca686a-7c72-4d8f-9290-e89635624a83",
		Name:           "Cavern of Souls",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		AsEnters:       ChooseCreatureTypeAsEnters("Cavern of Souls"),
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
				SpendRiders:      []game.ManaSpendRider{SpentSpellCantBeCountered(ManaRestrictCast)},
			},
		},
	})
}
