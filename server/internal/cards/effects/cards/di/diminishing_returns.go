package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diminishing Returns — Sorcery {2}{U}{U} (Eternal Masters, S82 OLD, #5542):
//
//	"Each player shuffles their hand and graveyard into their library. You
//	 exile the top ten cards of your library. Then each player draws up to
//	 seven cards."
//
// Time Reversal's wheel (every player's hand and graveyard go into their own
// library, and only once every shuffle has landed does the next step run) with
// a different ending: you exile the top ten cards of your library, face up,
// then each player in turn picks how many cards, 0 to seven, they draw (a draw
// limit lowers the ceiling, CR 121.2b; Fatal Lore's choose-a-number prompt).
// Diminishing Returns is on the stack while the graveyards are shuffled, so it
// is not among them, and it goes to the graveyard afterwards.
//
// No simplification.
func init() {
	effects.RegisterDiminishingReturnsTails()
	effects.Register(effects.Spec{
		OracleID:     "f9b02a7c-c122-4282-90e2-81fcfa663193",
		Name:         "Diminishing Returns",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.DiminishingReturnsResolve,
	})
}
