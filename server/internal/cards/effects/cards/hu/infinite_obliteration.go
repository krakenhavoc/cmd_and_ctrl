package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infinite Obliteration — Sorcery {1}{B}{B} (Magic Origins, S82 older sets):
//
//	"Choose a creature card name. Search target opponent's graveyard, hand,
//	 and library for any number of cards with that name and exile them. Then
//	 that player shuffles."
//
// The extraction family's search (Eradicate, Splinter) over a name its caster
// chooses as the spell resolves instead of one read off a card. The name is
// free text, so "creature card" is kept where it is read: only creature cards
// by that name are found, and a name no creature card has finds nothing (Day of
// the Moon). The caster searches all three zones in one prompt and may leave
// matches in the hand and library (CR 701.23b); the opponent shuffles
// whatever was taken. A target gone by resolution makes the spell do nothing
// and no name is asked (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bea62dc-14cd-4cbc-b55d-3791c444bdc7",
		Name:         "Infinite Obliteration",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.OldB321InfiniteObliteration("Infinite Obliteration"),
	})
}
