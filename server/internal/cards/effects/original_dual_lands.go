package effects

// original_dual_lands.go — the four Alpha "original duals" the
// roadmap's batch 02 (#295) ranks in the top 360:
//
//	Land — Island Swamp
//	"({T}: Add {U} or {B}.)"
//
// The whole card is its type line. There is no rules text at all —
// the parenthetical is reminder text for the mana abilities the two
// basic land types grant, which is exactly why these lands are what
// a fetchland wants and why they are still the best duals printed.
//
// They need a catalog entry anyway, and for the same reason the
// painlands did: the engine's synthetic land ability
// (ManaAbilitiesForCard → basicLandColor) only fires for lands
// carrying the BASIC supertype. An original dual is a NONBASIC land
// with basic land TYPES, so nothing derives its mana and an
// unregistered Underground Sea would sit on the battlefield
// producing nothing at all.
//
// The other half of the type line — being a real Island and a real
// Swamp — is carried by the printed TypeLine on the card itself and
// needs nothing here. That is what makes a Polluted Delta able to
// fetch one and a checkland able to see one, both of which already
// work through IsLandWithSubtype.
//
// One pipe ability rather than two single-colour entries, per
// dualManaAbility. IgnoreCommanderIdentity is not set and does not
// need to be: dualManaAbility leaves it false, and the two colours
// a dual produces are always inside the identity of a deck that can
// legally run it (CR 903.5d makes the land's own identity UB, so a
// deck playing Underground Sea has both colours in its commander's
// identity by construction).
//
// No simplifications — every word of these four cards is live.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"4b22be3a-8ce1-47d1-b82e-6c3ccfb0548b", "Underground Sea", "U", "B"},
		{"c718911c-c955-4eb9-9e16-be4bd49a4e4e", "Volcanic Island", "U", "R"},
		{"74b7fe23-5d3a-4092-8d78-7c0eba8f6f73", "Tropical Island", "G", "U"},
		{"02418479-9455-417f-a6a1-004356faff37", "Tundra", "W", "U"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
