package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bond_lands.go — the Battlebond / Commander Legends "bond lands":
//
//	"This land enters tapped unless you have two or more opponents."
//	"{T}: Add {A} or {B}."
//
// The most Commander-specific cycle in the format: printed for
// multiplayer, unplayable in a duel, and an untapped dual at every
// four-player table. Five of the top 165 cards on EDHREC.
//
// The first five landed with the top-100 staples batch (#261); the
// other five — every one of them in the roadmap's batch 01 (#294),
// ranks 167–210 — complete the cycle. Same helper, same shape, ten
// rows.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		// Top-100 staples batch (#261).
		{"bd004c9d-771e-4e63-a97d-a2259c096af8", "Morphic Pool", "U", "B"},
		{"d1620449-930a-4895-a143-fd2a0a3c8b17", "Rejuvenating Springs", "G", "U"},
		{"e3570ac7-c593-40e3-bbd6-ec3da6d8158d", "Training Center", "U", "R"},
		{"819e1765-8325-4e6f-89c1-63ea86de369f", "Luxury Suite", "B", "R"},
		{"672e190d-8ea0-4a2e-b74f-5d35304631e4", "Sea of Clouds", "W", "U"},
		// Roadmap batch 01 (#294).
		{"cf6d10ed-85c3-48f2-8ba0-2960e03b408b", "Spectator Seating", "R", "W"},
		{"ebc5ac83-08d4-4d6b-b840-0c4ba71a38ab", "Vault of Champions", "W", "B"},
		{"7c69f718-acc8-4851-8e5d-0cbaaa86192c", "Undergrowth Stadium", "B", "G"},
		{"45fe016e-1a09-410c-bbe3-4663ba06c5b7", "Spire Garden", "R", "G"},
		{"761cb262-f83b-4a99-9345-b773182a7671", "Bountiful Promenade", "G", "W"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(youHaveTwoOrMoreOpponents)},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
