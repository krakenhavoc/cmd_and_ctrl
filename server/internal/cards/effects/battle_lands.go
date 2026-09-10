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
// a Sunken Hollow. The engine's synthetic mana ability only fires for
// BASIC lands (ManaAbilitiesForCard → basicLandColor requires the
// "basic" supertype), so the pipe ability is declared here explicitly
// rather than left to fall out of the type line.
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
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(youControlTwoOrMoreBasics)},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
