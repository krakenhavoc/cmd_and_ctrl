package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of the Stampede — Creature — Human Knight {3}{G}, 2/4 (Rivals of
// Ixalan, #5531):
//
//	"Dinosaur spells you cast cost {2} less to cast."
//
// Dragonspeaker Shaman's reduction for Dinosaurs: a CostReduction over
// your Dinosaur spells, applied at CR 601.2f against generic mana only, so
// the coloured pips are never touched. An opponent's Dinosaur spell is not
// discounted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "801d60c1-40b1-48b6-8800-afe1a21386f3",
		Name:         "Knight of the Stampede",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Dinosaur spells you cast cost {2} less to cast.",
				YourSpell(), SpellOfCreatureType("Dinosaur")),
		},
	})
}
