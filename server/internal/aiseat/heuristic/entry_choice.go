package heuristic

// onlyLandInHandAndShort reports that the cards named are the seat's
// only land cards in hand and it controls fewer than three lands — the
// one case in which ADR 0098's Mox Diamond heuristic keeps the land and
// lets the Mox go to the graveyard.
func (st *state) onlyLandInHandAndShort(named []string) bool {
	if st.seat == nil {
		return false
	}
	handLands := 0
	for j := range st.seat.Hand.Cards {
		if isLand(&st.seat.Hand.Cards[j]) {
			handLands++
		}
	}
	if handLands > len(named) {
		return false
	}
	lands := 0
	for _, c := range st.bf {
		if c.Controller == st.me && isLand(c) {
			lands++
		}
	}
	return lands < 3
}
