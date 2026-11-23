package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borough Backup — {4}{W} Sorcery (Marvel Super Heroes, #4455):
//
//	"Create two 3/2 white Hero creature tokens with vigilance.
//	 Basic landcycling {2}"
//
// The tokens are one tokens_table template; basic landcycling is the
// cycling family's shared BasicLandcycling (CR 702.29).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6effbf59-3933-4a77-9a6a-fd3bedf45c26",
		Name:         "Borough Backup",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("3/2 white Hero with vigilance"), N: 2},
		),
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
