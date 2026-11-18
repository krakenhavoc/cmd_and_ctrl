package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patrician Geist — Creature — Spirit Knight {2}{U}, 2/2 (Innistrad: Double
// Feature):
//
//	"Flying
//	 Other Spirits you control get +1/+1.
//	 Spells you cast from your graveyard cost {1} less to cast."
//
// Legion Lieutenant's lord shape for Spirits (read through the layered
// view, so a changeling counts) and Emet-Selch's graveyard discount, which
// reads the zone the cast is coming from, so a flashback, an escape and a
// Snapcaster'd cast are all cheaper. The reduction spends generic mana only
// (CR 118.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1f65c853-5262-4b75-b324-1fe0c165170d",
		Name:            "Patrician Geist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Spirit"}, Others: true, YoursOnly: true}, 1, 1),
		},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Spells you cast from your graveyard cost {1} less to cast.",
				YourSpell(), CastFromGraveyard()),
		},
	})
}
