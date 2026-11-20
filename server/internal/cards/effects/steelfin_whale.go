package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steelfin Whale — Creature — Whale {5}{U}, 3/4 (Modern Horizons 2,
// #5501):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Whenever an artifact you control enters, untap this creature."
//
// Frogmite's affinity (CR 702.41a) and Skyswimmer Koi's trigger
// (neoB04ArtifactYouControlEntered) with Tiger-Seal's untap body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a35ff288-0ba0-47de-9848-4ff4bef6b91b",
		Name:         "Steelfin Whale",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, neoB04ArtifactYouControlEntered,
				"Steelfin Whale — untap this creature",
				func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
}
