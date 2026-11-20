package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timber Protector — Creature — Treefolk Warrior {4}{G}, 4/6 (Lorwyn,
// #5593):
//
//	"Other Treefolk creatures you control get +1/+1.
//	 Other Treefolk and Forests you control have indestructible."
//
// The first line is a lord (layer 7c, CR 613.4c). The second is a layer-6
// grant (CR 613.1f) over a wider set: every other Treefolk permanent you
// control, a Kindred one included, and every Forest you control, a nonbasic
// Forest and a land that is a Forest only through an effect included. The
// Protector itself is neither, so it has no indestructible of its own; an
// indestructible permanent isn't destroyed by lethal damage or a destroy
// effect (CR 702.12b).
//
// No simplification.
func init() {
	treefolk := lrwB1EachOtherYouControl("Treefolk")
	indestructible := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		if treefolk(target, g, source) {
			return true
		}
		return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
			target.IsLand() && target.HasSubtype(source.TextSubtype("Forest"))
	}, "indestructible")
	indestructible.Label = "Other Treefolk and Forests you control have indestructible."
	Register(Spec{
		OracleID:     "9bc8a69b-6065-4ef7-a580-adb1e7c72ef4",
		Name:         "Timber Protector",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Treefolk"}, Others: true, YoursOnly: true}, 1, 1),
			indestructible,
		},
	})
}
