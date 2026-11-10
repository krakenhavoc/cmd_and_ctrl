package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrumming Hivepool — Artifact {6} (Edge of Eternities):
//
//	"Affinity for Slivers (This spell costs {1} less to cast for each
//	 Sliver you control.)
//	 Slivers you control have double strike and haste.
//	 At the beginning of your upkeep, create two 1/1 colorless Sliver
//	 creature tokens."
//
// Affinity is a self cost modifier over Slivers you control (CR 702.41a),
// read through the layers so a changeling counts. The keyword grant is a
// layer 6 static over the Slivers the Hivepool's controller controls, by
// effective subtype; the Hivepool itself is an artifact and not one. The
// tokens are the printed Sliver (teoe 1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04ea21b3-084b-4746-9b49-99b6ef4f0885",
		Name:         "Thrumming Hivepool",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Slivers", HasSubtype("Sliver")),
		},
		Static: []game.StaticAbility{
			b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.HasSubtype(source.TextSubtype("Sliver"))
			}, "double strike", "haste"),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Thrumming Hivepool — create two 1/1 colorless Sliver creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 colorless Sliver"), N: 2})),
		},
	})
}
