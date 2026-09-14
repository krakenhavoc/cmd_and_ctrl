package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rocky Tar Pit — Land (EDHREC rank 3380):
//
//	"This land enters tapped.
//	 {T}, Sacrifice this land: Search your library for a Swamp or
//	 Mountain card, put it onto the battlefield, then shuffle."
//
// The Mirage fetchland for Rakdos — Mountain Valley's shape one
// colour pair over: enters tapped, and cracks for no life into any
// Swamp or Mountain card, nonbasics included as the printed "card"
// allows, which enters UNTAPPED (fetchDual, the Zendikar body without
// the life). The enters-tapped clause is the self-replacement, so the
// land is never seen untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8709b5b1-ef9e-45b2-bf4f-ef4c4d613dcd",
		Name:         "Rocky Tar Pit",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:  "{T}, Sacrifice this land: Search your library for a Swamp or Mountain card, put it onto the battlefield, then shuffle.",
			Cost:   Plus(TapCost(), SacrificeThis()),
			Effect: fetchDual("swamp", "mountain"),
		}},
	})
}
