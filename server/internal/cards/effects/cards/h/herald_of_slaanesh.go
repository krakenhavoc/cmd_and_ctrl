package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of Slaanesh — Creature — Demon {2}{R}, 2/2 (Warhammer 40,000
// Commander, S82 GAP batch 194, #5948):
//
//	"Locus of Slaanesh — Demon spells you cast cost {2} less to cast.
//	 Other Demons you control have haste."
//
// Goblin Warchief's two statics: the generic-only reduction on the Demon
// spells its controller casts (the Herald itself is not on the battlefield
// while it is cast, so it never discounts itself), and a layer 6 haste grant
// to the other Demons that player controls.
//
// No simplification.
func init() {
	haste := effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Demon"}, Others: true, YoursOnly: true}, "haste")
	haste.Label = "Other Demons you control have haste."
	effects.Register(effects.Spec{
		OracleID:     "ea912e4f-c734-4c7b-9e80-8961ca57cdd3",
		Name:         "Herald of Slaanesh",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Demon spells you cast cost {2} less to cast.", effects.YourSpell(), effects.SpellOfCreatureType("Demon")),
		},
		Static: []game.StaticAbility{haste},
	})
}
