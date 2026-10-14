package protocol

// stack_position_view.go: where each seat has dragged the stack panel to
// (#3429, ADR 0176).
//
// The placement is not part of the game. The room keeps it
// (Room.SetStackPosition), every capture stamps it on, and nothing about it
// enters the engine, its snapshot or its undo stack, exactly like the voice
// room and the rematch offer. The full view holds every seat's spot;
// FilterViewFor reduces it to the one viewer's own, so a seat's placement
// is never sent to anyone else.

// StackPositionView is the top-left corner of the stack panel as a fraction
// of the board's width and height, each in [0, 1]. The client clamps it so
// the panel stays fully on screen at whatever size the board has now.
type StackPositionView struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ValidStackPosition reports whether p is a placement the server stores:
// both fractions finite and within [0, 1].
func ValidStackPosition(p StackPositionView) bool {
	in := func(f float64) bool { return f >= 0 && f <= 1 } // false for NaN
	return in(p.X) && in(p.Y)
}

// filterStackPositionsFor reduces the full view's map to the viewer's own
// entry. A spectator has none and gets nil.
func filterStackPositionsFor(m map[string]StackPositionView, viewerID string) map[string]StackPositionView {
	if viewerID == SpectatorViewerID {
		return nil
	}
	p, ok := m[viewerID]
	if !ok {
		return nil
	}
	return map[string]StackPositionView{viewerID: p}
}
