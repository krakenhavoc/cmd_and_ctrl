package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "1f65c853-5262-4b75-b324-1fe0c165170d",
		Name:            "Patrician Geist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Spirit"}, Others: true, YoursOnly: true}, 1, 1),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Spells you cast from your graveyard cost {1} less to cast.",
				effects.YourSpell(), effects.CastFromGraveyard()),
		},
	})
}
