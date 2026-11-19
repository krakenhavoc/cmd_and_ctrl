package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akroma's Memorial — Legendary Artifact {7} (Special Guests, S82 BNS):
//
//	"Creatures you control have flying, first strike, vigilance, trample,
//	 haste, and protection from black and from red."
//
// One layer-6 grant (CR 613.1f) to every creature its controller controls,
// the source's own creature type irrelevant: a creature that arrives later
// has all of it at once, and one that changes control loses it. The two
// protections are two separate keywords, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "967e8b3f-8266-4d9a-a38b-d0e10184e1e4",
		Name:         "Akroma's Memorial",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "flying", "first strike", "vigilance",
				"trample", "haste", "protection from black", "protection from red"),
		},
	})
}
