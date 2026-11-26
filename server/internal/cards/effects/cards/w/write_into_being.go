package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Write into Being — {2}{U} Sorcery (Fate Reforged):
//
//	"Look at the top two cards of your library. Manifest one of those cards,
//	 then put the other on the top or bottom of your library."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fc69345-4c1f-4113-8e1c-15b8e2e597fc",
		Name:         "Write into Being",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.S82OldB337WriteIntoBeing{}),
	})
}
