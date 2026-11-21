package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Incarnation — Sorcery {8}{G} (Magic 2015, #5556):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Create three 3/3 green Beast creature tokens."
//
// Herd Migration's Beast token, three at once, with the convoke TapCost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31fefd48-6fff-41bf-a559-35a1699357ca",
		Name:         "Feral Incarnation",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 3}),
	})
}
