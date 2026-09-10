package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// check_lands.go — the ten Innistrad "checklands":
//
//	"This land enters tapped unless you control an <A> or a <B>."
//	"{T}: Add {A} or {B}."
//
// Ten of the top 130 cards on EDHREC, and the single largest block of
// the play-rate gap that needed no new machinery: the condition is an
// ordinary CR 614 self-replacement whose AppliesTo reads the board,
// and the dual is the Temple cycle's pipe ability.
//
// The check is on the land TYPE, not on a basic. In a real deck the
// thing that turns a checkland on is usually another nonbasic — a
// shockland, a fetched dual, a Triome — which is exactly why the
// cycle is played over the tapped duals and why the predicate must
// not be narrowed to basics.
//
// Grouped in one file for the same reason the Temples are: ten
// near-identical cards in ten files is ten places to fix the same
// mistake. The shape is stated once and each card is one row.
func init() {
	for _, land := range []struct{ oracleID, name, subA, subB, a, b string }{
		{"6a6c5e17-6465-4a1f-9d63-8a3ce2edc522", "Sulfur Falls", "island", "mountain", "U", "R"},
		{"d7faa3c8-46cf-46b2-bfa4-89000307cf18", "Clifftop Retreat", "mountain", "plains", "R", "W"},
		{"63398c02-6fb1-481d-9d9f-81063532fbc0", "Dragonskull Summit", "swamp", "mountain", "B", "R"},
		{"7e5d9efe-48a9-434b-bb09-056e0e09cc9a", "Isolated Chapel", "plains", "swamp", "W", "B"},
		{"027dd013-baa7-4111-b3c9-f4d1414e9c45", "Glacial Fortress", "plains", "island", "W", "U"},
		{"fb5a3403-7f0b-406c-8c4f-d693be010ca6", "Hinterland Harbor", "forest", "island", "G", "U"},
		{"819fc966-434e-470f-91e9-a38df974ad17", "Drowned Catacomb", "island", "swamp", "U", "B"},
		{"c9fe1383-1331-4a58-a45a-3320250221a9", "Woodland Cemetery", "swamp", "forest", "B", "G"},
		{"9516c4c1-d72d-434f-97e1-6a862434a169", "Rootbound Crag", "mountain", "forest", "R", "G"},
		{"402ec768-76fb-474e-ae74-babc90d833c4", "Sunpetal Grove", "forest", "plains", "G", "W"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(youControlLandTyped(land.subA, land.subB))},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
