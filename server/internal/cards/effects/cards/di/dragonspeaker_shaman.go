package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonspeaker Shaman — Creature — Human Barbarian Shaman {1}{R}{R}, 2/2:
//
//	"Dragon spells you cast cost {2} less to cast."
//
// Undead Warchief's reduction with a bigger number and no anthem: a
// CostReduction over "Dragon spells you cast" (YourSpell, SpellOfCreatureType),
// which the engine applies at CR 601.2f. The reduction spends against GENERIC
// mana only and stops at zero (game.reduceGeneric), so a {R}{R} Dragon still
// costs {R}{R} and a {2}{R}{R} one costs {R}{R}: the coloured pips are never
// touched. The type is read off the spell as it sits on the stack, so a
// Changeling is a Dragon spell (CR 702.73a), and an opponent's Dragon spell
// is not discounted. The Shaman is a Human Barbarian Shaman, not a Dragon, so
// it does not discount itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a29f36b-7911-4d84-a3ae-546b131d5159",
		Name:         "Dragonspeaker Shaman",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Dragon spells you cast cost {2} less to cast.",
				effects.YourSpell(), effects.SpellOfCreatureType("Dragon")),
		},
	})
}
