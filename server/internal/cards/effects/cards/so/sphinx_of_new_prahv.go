package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx of New Prahv — Creature — Sphinx {W}{W}{U}{U}, 4/3 (Ravnica
// Remastered, #5948):
//
//	"Flying, vigilance
//	 Spells your opponents cast that target this creature cost {2} more to
//	 cast."
//
// Kopala's tax, for this creature alone (ThatTargetThisPermanent): a cost increase (CR 601.2f) on a
// spell an opponent of the Sphinx's controller casts, read once the spell's
// targets are announced (CR 601.2c); a spell that also targets something else
// is taxed once. Abilities are not spells and are not taxed.
//
// No simplification.
func init() {
	spells := effects.CostsMore(2,
		"Spells your opponents cast that target this creature cost {2} more to cast.",
		effects.OpponentsSpell(), effects.ThatTargetThisPermanent())
	spells.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:        "3bdd3090-4503-4382-9a11-3e1cfcdffd71",
		Name:            "Sphinx of New Prahv",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		CostModifiers:   []game.CostModifier{spells},
	})
}
