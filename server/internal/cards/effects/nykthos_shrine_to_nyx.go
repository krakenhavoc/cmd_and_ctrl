package effects

// Nykthos, Shrine to Nyx — Legendary Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}: Choose a color. Add an amount of mana of that color
//	 equal to your devotion to that color."
//
// The batch-01 triage filed Nykthos under "choose a type", which is
// the wrong drawer twice over: the choice is a COLOUR, not a creature
// type, and it is made at activation rather than stored on the
// permanent. Nyx Lotus is the same sentence with a different cost and
// it shipped with #742, so this is that card's mana ability with a
// {2} in front of the tap.
//
// The whole ability is one colour pick whose AMOUNT depends on the
// colour picked: the ProducedFunc computes devotion to each colour at
// activation and writes it into the produced-mana grammar's
// per-option amounts — "{G4|U2}" is "four green or two blue". A
// colour with zero devotion is left off, because choosing it would
// add nothing, and with no devotion at all the ability adds nothing
// and the land still taps, as printed.
//
// Two abilities, not one with a mode. The colourless half is free and
// unconditional, so the auto-tapper can plan around it; the second
// costs {2} and is the deliberate click.
//
// Devotion counts the mana symbols in the printed costs of permanents
// the controller controls, hybrid symbols counting toward each of
// their colours — devotionTo, shared with Gray Merchant and Nyx
// Lotus, which is why the producer here is Nyx Lotus's own function
// rather than a second copy of it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84dc18f0-8225-4b40-a165-b10321e41769",
		Name:         "Nykthos, Shrine to Nyx",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true, Mana: "{2}"},
				ProducedFunc: nyxLotusProduced,
				Label:        "Choose a color: add mana of that color equal to your devotion to it",
			},
		},
	})
}
