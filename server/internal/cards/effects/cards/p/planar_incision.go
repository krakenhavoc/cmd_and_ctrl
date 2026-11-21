package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planar Incision — Instant {1}{U} (Kamigawa: Neon Dynasty, #5493):
//
//	"Exile target artifact or creature, then return it to the battlefield
//	 under its owner's control with a +1/+1 counter on it."
//
// Flicker with the Spaceshift "+1/+1 counter" rider, returning under the
// OWNER's control (a stolen creature goes home). The permanent comes back as
// a new object, entering with the counter; a token is exiled and stays gone
// (CR 111.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a959a91a-3e1d-4bd4-9e14-5ee977540146",
		Name:         "Planar Incision",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve:    effects.SocB22FlickerFirstTargetWithACounter,
	})
}
