package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thistledown Players — Creature — Mouse Bard {2}{W}, 3/3 (Bloomburrow,
// #4467):
//
//	"Whenever this creature attacks, untap target nonland permanent."
//
// Cat-Owl's attack trigger over any nonland permanent, yours or an
// opponent's (untapTheTarget). A target gone or no longer nonland by
// resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7054c42b-48dd-4887-b878-69c005259d27",
		Name:         "Thistledown Players",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Thistledown Players — untap target nonland permanent", untapTheTarget),
				TargetPermanent("target nonland permanent", Nonland())),
		},
	})
}
