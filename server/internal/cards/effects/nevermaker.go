package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nevermaker — Creature — Elemental {3}{U}, 2/3 (Morningtide, #5592):
//
//	"Flying
//	 When this creature leaves the battlefield, put target nonland
//	 permanent on top of its owner's library.
//	 Evoke {3}{U} (You may cast this spell for its evoke cost. If you do,
//	 it's sacrificed when it enters.)"
//
// Reveillark's leave trigger (any exit) with Lashweed Lurker's tuck of
// any nonland permanent, anyone's. Evoke is the alternative cost
// (CR 702.74a): cast for the same mana, the Nevermaker is sacrificed as
// it enters and the tuck still happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "38a2ef3e-3311-4e0a-807d-e5ccdd05af49",
		Name:             "Nevermaker",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Evoke("{3}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisLeaves("Nevermaker — put target nonland permanent on top of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					return putTargetOnTopOfOwnersLibrary(item, NewContext(g, item))
				}), TargetPermanent("target nonland permanent", Nonland())),
		},
	})
}
