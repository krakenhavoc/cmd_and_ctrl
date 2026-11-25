package game

// library_shuffle.go — the one shuffle of a library that a spell or
// ability causes, and its event (#6215; CR 701.24a: "To shuffle a
// library or a face-down pile of cards, randomize the cards within it
// so that no player knows their order").

// The Label of an EventLibraryShuffled.
const (
	// ShuffleByEffect: a spell or ability caused the shuffle.
	ShuffleByEffect = "effect"
	// ShuffleByAction: the sandbox's manual shuffle.
	ShuffleByAction = "action"
)

// shuffleLibraryLocked shuffles p's library from p's shuffle stream,
// un-knows every card in it, and announces it (EventLibraryShuffled
// with `label`). The event is emitted for an empty library too
// (CR 701.24e).
//
// Caller must hold g.mu (write).
func (g *Game) shuffleLibraryLocked(p *Player, label string) {
	if p == nil || p.Library == nil {
		return
	}
	p.Library.Shuffle(g.randForLocked(rngStream{kind: rngStreamShuffle, player: p.ID}))
	clearKnownInZoneLocked(p.Library)
	g.EmitEvent(Event{Kind: EventLibraryShuffled, Actor: p.ID, Target: p.ID, Label: label})
}
