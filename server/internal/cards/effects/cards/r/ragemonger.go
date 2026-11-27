package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ragemonger — {1}{B}{R} Creature — Minotaur Shaman 2/3 (Born of the Gods):
//
//	"Minotaur spells you cast cost {B}{R} less to cast. This effect reduces only the amount
//	 of colored mana you pay. (For example, if you cast a Minotaur spell with mana cost
//	 {2}{R}, it costs {2} to cast.)"
//
// Morophon's colours-only reduction (CR 118.7) with a {B}{R} unit: each symbol takes one
// mana of its colour off where the cost has it and nothing off generic mana. The spell's
// type is read through Ragemonger's text, so a text change reaches it (CR 612.2).
func init() {
	unit, err := game.ParseCost("{B}{R}")
	if err != nil {
		panic("effects.Ragemonger: " + err.Error())
	}
	reduction := effects.CostsLess(1, "Minotaur spells you cast cost {B}{R} less to cast. This effect reduces only the amount of colored mana you pay.",
		effects.YourSpell(), func(q game.CostQuery) bool { return q.Card.HasSubtype(q.Source.TextSubtype("Minotaur")) })
	reduction.Unit = &unit
	reduction.UnitColorsOnly = true
	effects.Register(effects.Spec{
		OracleID:      "0fcfc145-9371-40a8-9194-84e5fd700a8d",
		Name:          "Ragemonger",
		Completeness:  effects.CompletenessFull,
		CostModifiers: []game.CostModifier{reduction},
	})
}
