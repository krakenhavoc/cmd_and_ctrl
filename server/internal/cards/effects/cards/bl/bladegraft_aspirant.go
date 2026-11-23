package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladegraft Aspirant — Creature — Phyrexian Warrior {2}{R}, 2/3 (Phyrexia:
// All Will Be One):
//
//	"Menace
//	 Equipment spells you cast cost {1} less to cast.
//	 Activated abilities of Equipment you control that target this creature
//	 cost {1} less to activate."
//
// #5479, S82. Bureau Headmaster's Equipment-spell discount, and the
// Helitrooper discount read off the announced target, widened from equip
// abilities to any activated ability of an Equipment you control (the
// predicate reads the ability's source, not the equip bit). Like Helitrooper
// it declares ReadsTargets, so each legal target is priced on its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "644c785d-3061-44bc-9f78-bc939c565bf4",
		Name:            "Bladegraft Aspirant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Equipment spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.EquipmentSpell()),
			aspirantEquipmentAbilitiesThatTargetThis(),
		},
	})
}

// aspirantEquipmentAbilitiesThatTargetThis is "activated abilities of
// Equipment you control that target this creature cost {1} less to
// activate".
func aspirantEquipmentAbilitiesThatTargetThis() game.CostModifier {
	m := effects.ActivationCostsLess(1,
		"Activated abilities of Equipment you control that target this creature cost {1} less to activate.",
		func(q game.CostQuery) bool {
			return q.Card.HasSubtype("Equipment") && q.Card.Controller == q.Source.Controller
		}, effects.YouActivate(), effects.ThatTargetThisPermanent())
	m.ReadsTargets = true
	return m
}
