package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rags // Riches — Sorcery // Sorcery, Aftermath (Amonkhet, S82 OLD batch 243, #5535):
//
//	Rags {2}{B}{B}: "All creatures get -2/-2 until end of turn."
//	Riches {5}{U}{U}: "Aftermath. Each opponent chooses a creature they
//	 control. You gain control of those creatures."
//
// Two registered faces. Rags is Biting Rain's -2/-2 on every creature, yours
// included. Riches asks each opponent that controls a creature, in turn order,
// for one of theirs (CR 101.4); once every answer is in, you gain control of
// each chosen creature for good (CR 611.2a). It is cast only from a graveyard
// and exiled as it resolves (CR 702.127a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2c66419-c2ca-4259-afc6-eaec6d3bb897",
		Name:         "Rags",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		OnResolve:    effects.ResolveSteps(effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Toughness: -2, Label: "Rags — -2/-2 until end of turn"}),
	})
	effects.Register(effects.Spec{
		OracleID:     "c2c66419-c2ca-4259-afc6-eaec6d3bb897#1",
		Name:         "Riches",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB243EachOpponentChoosesACreatureYouGainControlOfIt("Riches"),
	})
}
