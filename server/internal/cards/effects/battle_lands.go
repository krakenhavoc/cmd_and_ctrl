package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// battle_lands.go — the five Battle for Zendikar "battle lands":
//
//	"This land enters tapped unless you control two or more basic lands."
//	"({T}: Add {A} or {B}.)"
//
// The mana ability is in reminder-text parentheses on the printed
// card because it comes from the land types themselves — these really
// are "Land — Island Swamp", which is why a Polluted Delta can fetch
// a Sunken Hollow. The engine's synthetic mana ability used to fire
// only for BASIC lands, so the pipe ability was declared here
// explicitly rather than left to fall out of the type line. That
// shortcut is gone: ManaAbilitiesForCard now derives the CR 305.6
// intrinsic abilities from a land's effective subtypes, supertype
// or not. The declarations below are kept because a single pipe
// ("Add {U} or {B}") is one click in the client where two separate
// abilities are two, and because ManaAbilitiesForCard adds only the
// colours a declaration cannot already make — so these cards get
// exactly the abilities they had, and an Urborg still adds {B} to
// the ones that can't produce it.
//
// The condition counts BASICS, unlike the checklands' land-type
// check. That is the whole difference between the two cycles and it
// runs the opposite way in deck construction: a checkland wants
// nonbasic duals, a battle land wants a basics-heavy mana base.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"cd2c90ac-2b04-461c-92f3-939871b6b6a3", "Sunken Hollow", "U", "B"},
		{"dfac0258-e148-4d7d-8ded-fc2466d9caa6", "Cinder Glade", "R", "G"},
		{"390f1b56-264e-4336-83be-dc1fe79bfdcf", "Smoldering Marsh", "B", "R"},
		{"dcb7e046-f01b-497c-88e5-57794eb30ce5", "Canopy Vista", "G", "W"},
		{"5330e24a-8568-446e-840a-594cd08bd1bc", "Prairie Stream", "W", "U"},
		// Roadmap batch 06 (#299): the enemy-pair printing of the same
		// two-basics clause.
		{"40544d12-0391-4a61-af95-9b8ec01ed8fc", "Vernal Fen", "B", "G"},
		// Roadmap batch 10 (#303): Land — Mountain Plains, the same
		// two-basics clause.
		{"5dd0cc44-4647-4857-ad3b-22494099d08a", "Radiant Summit", "R", "W"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(youControlTwoOrMoreBasics)},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
