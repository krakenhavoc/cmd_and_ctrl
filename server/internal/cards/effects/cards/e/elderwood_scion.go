package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elderwood Scion — {3}{G}{W} Creature — Elemental 4/4:
//
//	"Trample, lifelink
//	 Spells you cast that target this creature cost {2} less to cast.
//	 Spells your opponents cast that target this creature cost {2} more to
//	 cast."
//
// Boreal Elemental's tax and its mirror image, two cost modifiers that read
// the announced targets (CR 601.2c, 601.2f). A reduction spends against
// generic mana only. Abilities are not spells and are not touched.
func init() {
	less := effects.CostsLess(2,
		"Spells you cast that target this creature cost {2} less to cast.",
		effects.YourSpell(), effects.ThatTargetThisPermanent())
	less.ReadsTargets = true
	more := effects.CostsMore(2,
		"Spells your opponents cast that target this creature cost {2} more to cast.",
		effects.OpponentsSpell(), effects.ThatTargetThisPermanent())
	more.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:        "99c8647f-e08d-4217-b637-3f8f2bffdaa3",
		Name:            "Elderwood Scion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "lifelink"},
		CostModifiers:   []game.CostModifier{less, more},
	})
}
