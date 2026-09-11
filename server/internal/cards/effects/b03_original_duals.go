package effects

// original_duals.go — the Alpha dual lands the roadmap's batch 03
// (#296) ranks: Badlands, Scrubland, Bayou, Taiga, Plateau, Savannah.
//
//	"Land — Swamp Mountain
//	 ({T}: Add {B} or {R}.)"
//
// The entire text is reminder text for the two basic land types, and
// that is exactly why a catalog entry is needed: the engine's
// synthetic land ability (ManaAbilitiesForCard → basicLandColor)
// derives mana only for a land with the BASIC supertype, so an
// unregistered Badlands sat on the battlefield producing nothing —
// the same "worse than absent" hole the painlands and battle lands
// had. The printed type line is kept, so a checkland reading land
// types (Dragonskull Summit off a Badlands) and a fetchland's
// "Swamp card" both see the dual as printed.
//
// No enters-tapped, no drawback: one dualManaAbility row each, the
// same shape the bond and check lands use. The pipe narrows to the
// controller's commander identity at activation, as every two-colour
// land in the catalog does — a UX narrowing that removes an option
// the deck could never use and never adds one.
func init() {
	for _, d := range []struct{ oracleID, name, a, b string }{
		{"13ff3222-91cb-4796-a34e-899ed817694c", "Badlands", "B", "R"},
		{"c8d95ca8-7d12-4072-aeaf-e20f248c7e39", "Scrubland", "W", "B"},
		{"b76d1ae6-ad1d-4bac-b4c3-2e03e0e84d9b", "Bayou", "B", "G"},
		{"22e3cf1d-3559-4ce1-954c-8dc815342979", "Taiga", "R", "G"},
		{"c7a15ca4-085f-4d92-8387-c3711c04c8fa", "Plateau", "R", "W"},
		{"703243f0-8cb3-420f-958f-5fd4bde30293", "Savannah", "G", "W"},
	} {
		Register(Spec{
			OracleID:      d.oracleID,
			Name:          d.name,
			ManaAbilities: []ManaAbility{dualManaAbility(d.a, d.b)},
		})
	}
}
