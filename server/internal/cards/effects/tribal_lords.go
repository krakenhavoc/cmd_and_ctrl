package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tribal_lords.go — the five S26 lords, registered from one table.
//
// Written as a table rather than five files for the reason the Temple
// cycle is: they differ only in a noun, a colour and whether the
// clause says "you control", and five near-identical files are five
// places to make the same mistake. The one lord with an extra ability
// (Elvish Archdruid's mana) declares it in the same entry, since
// Spec is the whole card either way.
//
// The distinction the table makes visible, and the reason the
// builders take it as a field: HALF THESE LORDS PUMP THE WHOLE
// TABLE. Lord of Atlantis, Elvish Champion and Goblin King say
// "Other X get +1/+1" with no controller clause — they buff an
// opponent's Elves too, and playing two Goblin Kings across the table
// from each other is a real and printed outcome. Goblin Chieftain and
// Elvish Archdruid say "you control" and do not.
//
// LANDWALK IS GRANTED SINCE #705. Elvish Champion's forestwalk and
// Goblin King's mountainwalk were left out in S26, because the
// block check could not see the defending player's lands and a badge
// for an evasion the engine did not provide was worse than no badge.
// The check is game-aware now (game/landwalk.go), so both lords grant
// their landwalk, symmetrically like their anthems, and are complete.

func init() {
	for _, lord := range []struct {
		oracleID string
		name     string
		filter   TribeFilter
		keyword  string
		spec     func(*Spec)
	}{
		{
			// "Other Elf creatures you control get +1/+1.
			//  {T}: Add {G} for each Elf you control."
			// The mana half counts EVERY Elf you control, itself
			// included — the ability does not say "other" — and it
			// counts changelings, which is what makes an Elf deck's
			// Universal Automaton a real ramp piece.
			oracleID: "6e2c2423-d854-4478-99e6-64f29851f026",
			name:     "Elvish Archdruid",
			filter:   TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true},
			spec: func(s *Spec) {
				s.ManaAbilities = []ManaAbility{{
					Cost:         ManaAbilityCost{Tap: true},
					ProducedFunc: ProducedPerPermanent("G", MatchCreatureSubtype("Elf")),
					Label:        "Add {G} for each Elf you control",
				}}
			},
		},
		{
			// "Other Elf creatures get +1/+1 and have forestwalk."
			// No "you control" — every Elf on the battlefield.
			oracleID: "7e40f37c-9a0c-40e0-b195-7ea94b12f798",
			name:     "Elvish Champion",
			filter:   TribeFilter{Tribes: []string{"Elf"}, Others: true},
			keyword:  "forestwalk",
			spec:     func(s *Spec) { s.Completeness = CompletenessFull },
		},
		{
			// "Other Goblins get +1/+1 and have mountainwalk."
			oracleID: "d236b3fc-0d3f-4d99-875d-e32a33fe5767",
			name:     "Goblin King",
			filter:   TribeFilter{Tribes: []string{"Goblin"}, Others: true},
			keyword:  "mountainwalk",
			spec:     func(s *Spec) { s.Completeness = CompletenessFull },
		},
		{
			// "Haste. Other Goblin creatures you control get +1/+1
			//  and have haste."
			// Its own haste is printed data off the Scryfall keyword
			// array, so only the GRANT is declared here.
			oracleID: "368b4052-174e-4458-a6e6-eaf8093aa0fe",
			name:     "Goblin Chieftain",
			filter:   TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true},
			keyword:  "haste",
			spec:     func(s *Spec) { s.PrintedKeywords = []string{"haste"} },
		},
	} {
		spec := Spec{
			OracleID: lord.oracleID,
			Name:     lord.name,
			Static:   []game.StaticAbility{TribalAnthem(lord.filter, 1, 1)},
		}
		if lord.keyword != "" {
			spec.Static = append(spec.Static, TribalKeywordGrant(lord.filter, lord.keyword))
		}
		if lord.spec != nil {
			lord.spec(&spec)
		}
		Register(spec)
	}
}
