package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shatterstorm — Sorcery {2}{R}{R}:
//
//	"Destroy all artifacts. They can't be regenerated."
//
// The unconditional artifact wrath, and the answer to the mana-rock
// pile every Commander deck is built on. Symmetric, unlike
// Vandalblast's overload, which is why a red deck plays whichever
// one matches its own board: Vandalblast if you have rocks of your
// own, Shatterstorm if you do not and want them all gone.
//
// "They can't be regenerated" is ENFORCED as of #667
// (CR 701.19c). See mass.go.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96ce2403-4607-440a-92ae-80aceb458c5d",
		Name:         "Shatterstorm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Artifact(), CantBeRegenerated: true}.Apply(ctx)
		},
	})
}
