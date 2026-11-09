package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b72b_helpers.go — shared pieces for Lorwyn Eclipsed batch 72b
// (#4458).

// mutavaultOracleID is the card Mutavault's oracle ID (mutavault.go).
const mutavaultOracleID = "6b3cc59a-7ea5-4eb5-9bf9-5a9c07f80e2b"

// MutavaultToken is "a Mutavault token": a token created by name, with
// no other characteristics given, takes the characteristics of the card
// with that name in the Oracle card reference (CR 111.11; Mutable
// Explorer's ruling says the same). So the token carries Mutavault's
// oracle ID, exactly as a token copy carries the copied card's, and its
// "{T}: Add {C}" and "{1}: becomes a 2/2 with all creature types" are the
// catalog card's own rows. A colourless Token Land, as printed (tecl 11).
func MutavaultToken() game.Card {
	return game.Card{
		Name:         "Mutavault",
		OracleID:     mutavaultOracleID,
		TypeLine:     "Token Land",
		ProducedMana: []string{"C"},
	}
}
