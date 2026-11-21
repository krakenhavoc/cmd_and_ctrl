package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Command Performance — Sorcery {1}{U}:
//
//	"Choose two —
//	 • Open an Attraction. (Put the top card of your Attraction deck onto
//	   the battlefield.)
//	 • Roll to visit your Attractions.
//	 • You get {TK}{TK}.
//	 • You may put a sticker on a nonland permanent you own."
//
// The bullets are performed in printed order (CR 700.2): the open and the
// roll are the engine's (ADR 0207), the tickets a player counter (CR
// 107.17) and the sticker the engine's keyed questions (ADR 0208 §4), so
// "get {TK}{TK}" and "put a sticker" chosen together pay for the sticker
// with the tickets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6c079cd-69c2-4da4-ba41-91844a4d2f06",
		Name:         "Command Performance",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeSteps("Open an Attraction.", nil, effects.OpenAttractions{N: 1}),
			effects.ModeSteps("Roll to visit your Attractions.", nil, effects.RollToVisitYourAttractions{}),
			effects.ModeSteps("You get {TK}{TK}.", nil, effects.GetTickets{N: 2}),
			effects.ModeSteps("You may put a sticker on a nonland permanent you own.", nil, effects.PutASticker{}),
		),
	})
}
