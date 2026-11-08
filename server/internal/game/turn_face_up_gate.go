package game

// turn_face_up_gate.go — #5241: "… can't be turned face up" (CR 101.2).
//
// Karlov Watchdog's "Permanents your opponents control can't be turned
// face up during your turn" is a static ability of a permanent on the
// battlefield, and a "can't" beats every "may" (CR 101.2). Its ruling
// names the doors it shuts: the CR 116.2b special action, whether it
// pays a morph or disguise cost (CR 702.37e, CR 702.168d) or a
// manifested or cloaked creature card's mana cost (CR 701.40b,
// CR 701.58b). "Can't be turned face up" is about the permanent, not
// about the action, so an effect that would turn it face up (CR 708.8)
// does nothing to it either.
//
// The shape is LandPlayRestriction's (ADR 0109 §4): a struct of hooks
// on CardDef, read from the battlefield through CatalogAbilityKey, so
// nothing is stored, the source leaving lifts the ban, and a source
// that has lost its abilities (CR 613.1f) bans nothing.
//
// TWO CALLERS, ONE FUNCTION:
//
//   - SpecialActionTimingOKLocked's turn_face_up row, which the engine,
//     the legal-move enumerator and the view all ask, so the action is
//     refused, never offered to a bot, and greyed on the client;
//   - TurnFaceUpForEffect, the one door an effect uses (and
//     TurnFaceUpForbiddenForEffect, which a card asks before it offers
//     a turn that would do nothing).

// TurnFaceUpQuery is everything a turn-face-up restriction may look at.
type TurnFaceUpQuery struct {
	// Game is the game being asked about. Read-only.
	Game *Game
	// Card is the face-down permanent that would be turned face up.
	Card *Card
	// Source is the permanent contributing the restriction; its
	// Controller is the "you" of the ability.
	Source *Card
}

// TurnFaceUpRestriction is one "can't be turned face up" static
// ability contributed by a permanent on the battlefield (#5241).
type TurnFaceUpRestriction struct {
	// Label is the clause as printed.
	Label string
	// Forbids decides whether this restriction stops q.Card being
	// turned face up right now. Nil forbids nothing. Read-only, under
	// g.mu.
	Forbids func(q TurnFaceUpQuery) bool
}

// CatalogTurnFaceUpRestrictions is the catalog hook, wired from CardDef
// at init like CatalogLandPlayRestrictions.
var CatalogTurnFaceUpRestrictions func(oracleID string) []TurnFaceUpRestriction

// turnFaceUpForbiddenLocked reports whether a static ability on the
// battlefield stops the permanent `c` being turned face up right now.
//
// Caller must hold g.mu (read or write).
func (g *Game) turnFaceUpForbiddenLocked(c *Card) bool {
	if g.Battlefield == nil || CatalogTurnFaceUpRestrictions == nil || c == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, r := range CatalogTurnFaceUpRestrictions(key) {
			if r.Forbids != nil && r.Forbids(TurnFaceUpQuery{Game: g, Card: c, Source: src}) {
				return true
			}
		}
	}
	return false
}

// TurnFaceUpForbiddenForEffect is turnFaceUpForbiddenLocked for the
// effects package: a card that offers a choice to turn a permanent face
// up asks it first, so it never offers a turn that would do nothing.
//
// The *ForEffect surface: caller must hold g.mu.
func (g *Game) TurnFaceUpForbiddenForEffect(c Card) bool {
	return g.turnFaceUpForbiddenLocked(&c)
}
