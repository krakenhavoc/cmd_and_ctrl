package effects

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
	Register(Spec{
		OracleID:     "f6c079cd-69c2-4da4-ba41-91844a4d2f06",
		Name:         "Command Performance",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeSteps("Open an Attraction.", nil, OpenAttractions{N: 1}),
			ModeSteps("Roll to visit your Attractions.", nil, RollToVisitYourAttractions{}),
			ModeSteps("You get {TK}{TK}.", nil, GetTickets{N: 2}),
			ModeSteps("You may put a sticker on a nonland permanent you own.", nil, PutASticker{}),
		),
	})
}
