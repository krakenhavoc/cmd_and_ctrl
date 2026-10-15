package protocol

// pass_hint.go — #3826: TurnView.pass_hint, what the viewer's own next
// pass hands over.
//
// The action dock's next button looks the same at every priority pass,
// but two of them are not routine: the last one in your own turn (a
// pass there ends it) and the last one before your turn begins (your
// next decision is in your own turn). The engine decides both, so
// neither a client nor a model has to read the turn order and the
// plan off the board and hope it agrees:
//
//   - the plan (TurnView.upcoming) says whether any step that grants
//     priority is left this turn, extra phases and added steps included;
//   - TurnView.next_seat (ADR 0188 Decision 9) says whose turn follows,
//     with extra turns, skip charges and departed seats already counted.
//
// What it does not see: the viewer's own auto-pass and stops. Those are
// the client's (ADR 0009), so the hint is the structural fact, "no later
// step of this turn will give you priority", and never a promise about
// which earlier window the client stops at. A client that already
// auto-passed to the end step simply shows the hint there. It is also
// about the viewer's own pass: an opponent can still answer, and then
// the viewer is asked again.

// The values of TurnView.PassHint. An empty hint means "a routine pass".
const (
	// PassHintEndsTurn: the viewer is the active player, holds priority
	// on an empty stack, and no later step of this turn grants priority,
	// so passing hands the turn over.
	PassHintEndsTurn = "ends_turn"
	// PassHintYourTurnNext: it is another player's turn, the viewer holds
	// priority on an empty stack with no later priority step in it, and
	// the turn after this one is the viewer's, so their next decision is
	// in their own turn.
	PassHintYourTurnNext = "your_turn_next"
)

// passHintFor is the hint for the seat viewerID, or "". It reads only
// public turn state, so it tells the viewer nothing the table could not
// work out, and it is stamped per viewer because "ends your turn" and
// "your turn next" are about the viewer.
func passHintFor(v GameView, viewerID string) string {
	if viewerID == "" || viewerID == SpectatorViewerID {
		return ""
	}
	if v.State != "active" || v.MulligansOpen || v.OpeningRoll != nil || v.Turn.BetweenTurns {
		return ""
	}
	seat := -1
	for i, p := range v.Seats {
		if p.ID == viewerID {
			if p.Eliminated {
				return ""
			}
			seat = i
			break
		}
	}
	if seat < 0 || v.Turn.PriorityHolder != seat {
		return ""
	}
	// A pass with something on the stack resolves it and asks again.
	if len(v.Stack.Cards) > 0 {
		return ""
	}
	// Untap and cleanup grant no priority (CR 502.4, 514.3), so a
	// holder there is not a window the hint can speak for.
	if v.Turn.Step == "untap" || v.Turn.Step == "cleanup" {
		return ""
	}
	for _, p := range v.Turn.Upcoming {
		if p.Step != "cleanup" && p.Step != "untap" {
			return ""
		}
	}
	if v.Turn.ActiveSeat == seat {
		return PassHintEndsTurn
	}
	if v.Turn.NextSeat != nil && *v.Turn.NextSeat == seat {
		return PassHintYourTurnNext
	}
	return ""
}

// turnViewFor is the turn as viewerID sees it: the public turn, plus the
// viewer's own pass hint (#3826). A previous filter's hint is replaced,
// never kept, so filtering twice for different viewers is safe.
func turnViewFor(v GameView, viewerID string) TurnView {
	t := v.Turn
	t.PassHint = passHintFor(v, viewerID)
	return t
}
