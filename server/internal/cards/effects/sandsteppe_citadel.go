package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandsteppe Citadel — Land (EDHREC rank 506):
//
//	"This land enters tapped.
//	 {T}: Add {W}, {B}, or {G}."
//
// The Khans of Tarkir tri-land: a Temple without the scry and with a
// third colour. Enters-tapped is the CR 614 self-replacement; the
// mana is a three-way pipe over the printed colours, commander
// identity listed first, like every pipe whose text says nothing
// about the command zone.
//
// One card file rather than a cycle table because batch 03 is
// writing the other tri-lands concurrently; fold this row into that
// table when both have landed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "544dbabd-cbfc-40da-a5ba-2fea9cddb453",
		Name:         "Sandsteppe Citadel",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|B|G}",
			Label:    "Add {W}, {B}, or {G}",
		}},
	})
}
