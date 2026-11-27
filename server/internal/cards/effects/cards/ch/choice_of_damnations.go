package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Choice of Damnations — Sorcery — Arcane {5}{B} (Saviors of Kamigawa,
// #5601):
//
//	"Target opponent chooses a number. You may have that player lose that
//	 much life. If you don't, that player sacrifices all but that many
//	 permanents of their choice."
//
// The opponent names any number from 0 up (CR 107.1b has no printed ceiling).
// The caster is then asked whether that player loses that much life; on a no
// the opponent chooses which of their permanents, lands and tokens included,
// to keep (that many), and every other one is sacrificed at once (CR 701.21).
// An opponent with that many permanents or fewer keeps them all; a chosen 0
// sacrifices everything with nothing to choose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "506714a8-b9c2-443a-a9e8-ef9130e674ba",
		Name:         "Choice of Damnations",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepSacrifice, OpponentsOnly: true, Partial: true}},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.OldB586ChoiceOfDamnations,
	})
}
