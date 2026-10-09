package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radiant Destiny — Enchantment {2}{W} (EDHREC rank 5609):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 As this enchantment enters, choose a creature type.
//	 Creatures you control of the chosen type get +1/+1. As long as you
//	 have the city's blessing, they also have vigilance."
//
// Vanquisher's Banner's anthem with a keyword on top. The choice is the
// CR 614.12 as-enters prompt the tribal permanents share, answered
// before the permanent is on the battlefield, so the anthem never reads
// an empty type; the two statics read the type back off the permanent
// (TribeFilter{Chosen: true}) and match nothing until it is answered.
//
// The vigilance half is a layer-6 grant gated on the controller's
// blessing, and the layer pass is invalidated when the blessing is
// earned, so the creatures untap-attack the turn it happens. Once
// earned it is never lost (CR 702.131c).
//
// No simplification.
func init() {
	chosen := TribeFilter{Chosen: true, YoursOnly: true}
	Register(Spec{
		OracleID:        "068cbbe3-43c3-425e-80ec-d1ce5c02ea9b",
		Name:            "Radiant Destiny",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		AsEnters:        ChooseCreatureTypeAsEnters("Radiant Destiny"),
		Static: []game.StaticAbility{
			TribalAnthem(chosen, 1, 1),
			KeywordGrant(WhileCitysBlessing(chosen.Matches), "vigilance"),
		},
	})
}
