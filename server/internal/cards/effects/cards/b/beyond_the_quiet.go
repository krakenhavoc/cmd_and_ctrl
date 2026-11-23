package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beyond the Quiet — Sorcery {3}{W}{W} (Edge of Eternities, #4461):
//
//	"Exile all creatures and Spacecraft."
//
// Farewell's exile sweep (ExileAllMatching) over every creature and
// every Spacecraft, whoever controls them: a Spacecraft below its
// station threshold is not a creature and still goes, and so does a
// creature that is no Spacecraft. The set is read once as the spell
// resolves. Exile is not destruction, so indestructible does not save
// anything.
//
// The sweep is declared as nonland permanents, partial: it spares
// noncreature permanents that are not Spacecraft (ADR 0126 §6, the
// smallest listed class that holds everything it removes).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c96424e7-63c4-4f07-959f-6ad09f91279a",
		Name:         "Beyond the Quiet",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepExile, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileAllMatching{Match: effects.Or(effects.Creature(), effects.OfSubtype("Spacecraft"))}.Apply(ctx)
		},
	})
}
