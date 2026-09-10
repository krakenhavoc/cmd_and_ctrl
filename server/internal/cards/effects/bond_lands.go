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
// Five of the ten are in the top 100 by play rate; the other five
// (Bountiful Promenade, Spire Garden, Vault of Champions, Sacred
// Peaks, Undergrowth Stadium) sit just outside it and are deliberately
// left for the next batch rather than smuggled in — this file's list
// is the play-rate cut, not the cycle.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"bd004c9d-771e-4e63-a97d-a2259c096af8", "Morphic Pool", "U", "B"},
		{"d1620449-930a-4895-a143-fd2a0a3c8b17", "Rejuvenating Springs", "G", "U"},
		{"e3570ac7-c593-40e3-bbd6-ec3da6d8158d", "Training Center", "U", "R"},
		{"819e1765-8325-4e6f-89c1-63ea86de369f", "Luxury Suite", "B", "R"},
		{"672e190d-8ea0-4a2e-b74f-5d35304631e4", "Sea of Clouds", "W", "U"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(youHaveTwoOrMoreOpponents)},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
