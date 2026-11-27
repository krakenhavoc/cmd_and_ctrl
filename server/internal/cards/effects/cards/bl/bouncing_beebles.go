package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bouncing Beebles — Creature — Beeble {2}{U} 2/2:
//
//	"This creature can't be blocked as long as defending player controls an
//	 artifact."
//
// Read as each block is declared (CR 509.1b), against the artifacts the
// defending player controls then.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fa46972-d953-4328-9c31-96e134234133",
		Name:         "Bouncing Beebles",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.OldB719CantBeBlockedWhileDefenderControls(effects.Artifact())},
	})
}
