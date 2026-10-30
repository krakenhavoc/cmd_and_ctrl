package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6effbf59-3933-4a77-9a6a-fd3bedf45c26",
		Name:         "Borough Backup",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("3/2 white Hero with vigilance"), N: 2},
		),
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
