package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embereth Shieldbreaker // Battle Display — Creature — Human Knight {1}{R}, 2/1 // Sorcery — Adventure {R} (Throne of Eldraine, #5516):
//
//	Embereth Shieldbreaker — no rules text.
//	Battle Display — "Destroy target artifact."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const emberethShieldbreakerOracleID = "52e439d0-a263-4d53-9ae3-fa1f8aa8293e"

func init() {
	effects.Register(effects.Spec{
		OracleID:     emberethShieldbreakerOracleID,
		Name:         "Embereth Shieldbreaker",
		Completeness: effects.CompletenessFull,
	})
	effects.Register(effects.Spec{
		OracleID:     emberethShieldbreakerOracleID + "#1",
		Name:         "Battle Display",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
