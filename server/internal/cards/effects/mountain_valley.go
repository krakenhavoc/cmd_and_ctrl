package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mountain Valley — Land (EDHREC rank 3130):
//
//	"This land enters tapped.
//	 {T}, Sacrifice this land: Search your library for a Mountain or
//	 Forest card, put it onto the battlefield, then shuffle."
//
// The Mirage fetchland for Gruul: enters tapped, and cracks for no
// life into any Mountain or Forest card — nonbasics included, as
// the printed "card" allows — which enters UNTAPPED (fetchDual, the
// Zendikar body without the life). The enters-tapped clause is the
// self-replacement, so the land is never seen untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b7393aa-d563-45bc-9946-8e7d1729d498",
		Name:         "Mountain Valley",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:  "{T}, Sacrifice this land: Search your library for a Mountain or Forest card, put it onto the battlefield, then shuffle.",
			Cost:   Plus(TapCost(), SacrificeThis()),
			Effect: fetchDual("mountain", "forest"),
		}},
	})
}
