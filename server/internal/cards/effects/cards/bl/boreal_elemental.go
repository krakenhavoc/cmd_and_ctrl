package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boreal Elemental — Creature — Elemental {4}{U}, 3/4 (Core Set 2020, #5518):
//
//	"Flying
//	 Spells your opponents cast that target this creature cost {2} more to
//	 cast."
//
// Sphinx of New Prahv's tax: a cost increase (CR 601.2f) on a spell an
// opponent of the Elemental's controller casts, read once the spell's targets
// are announced (CR 601.2c); a spell that also targets something else is taxed
// once. Abilities are not spells and are not taxed.
//
// No simplification.
func init() {
	spells := effects.CostsMore(2,
		"Spells your opponents cast that target this creature cost {2} more to cast.",
		effects.OpponentsSpell(), effects.ThatTargetThisPermanent())
	spells.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:        "68a2e371-ccf0-4c2a-aeef-c66852811e1a",
		Name:            "Boreal Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers:   []game.CostModifier{spells},
	})
}
