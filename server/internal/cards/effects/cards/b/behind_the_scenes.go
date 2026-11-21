package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Behind the Scenes — Enchantment {2}{B} (Tarkir: Dragonstorm Commander):
//
//	"Creatures you control have skulk. (They can't be blocked by
//	 creatures with greater power.)
//	 {4}{W}: Creatures you control get +1/+1 until end of turn."
//
// The skulk is a layer-6 grant to every creature its controller has
// (the tribal grant with an empty tribe list); the block rule itself is
// the engine's (CR 702.118). The pump is Dawnwing Marshal's: it reaches
// the creatures you control as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e84837bd-4f94-43e1-876e-d82cae6ee247",
		Name:         "Behind the Scenes",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "skulk")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    effects.ManaCost("{4}{W}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Effect: effects.Do(effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
				Label: "Behind the Scenes — creatures you control get +1/+1",
			}),
		}},
	})
}
