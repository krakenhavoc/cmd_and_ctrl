package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syr Faren, the Hengehammer — Legendary Creature — Human Knight {G}{G}, 2/2
// (Throne of Eldraine, S82 old sets):
//
//	"Whenever Syr Faren attacks, another target attacking creature gets +X/+X
//	 until end of turn, where X is Syr Faren's power."
//
// Wildwood Mentor's attack trigger (OldB147AnotherAttackerGetsXX): the target
// is chosen as it goes on the stack, X is Syr Faren's power as it resolves (or
// its last-known power), and "another" excludes Syr Faren himself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2158fb0d-c199-4cb1-848c-077113348980",
		Name:         "Syr Faren, the Hengehammer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB147AnotherAttackerGetsXX("Syr Faren — another target attacking creature gets +X/+X until end of turn, where X is his power"),
		},
	})
}
