package game

import "slices"

// snapshot_backfill.go recomputes fields that a restore point written
// by an older binary does not carry, where the zero value would be
// wrong and the new binary can work out the right one. See the rule on
// SnapshotSchemaVersion.

// PrintedVariableToughness reports whether a printing's printed
// toughness is not a number ("*", "1+*", "?"). The deck importer asks
// the same question when it stamps Card.VariableToughness. top is the
// answer for the top-level toughness, faces has one answer per card
// face (nil for a single-faced printing), and ok is false when this
// process does not know the printing.
//
// cmd/server installs it from the Scryfall index
// (deck.PrintedVariableToughness) before it restores any game. Nil
// means there is no index: the game package's own tests, or a server
// with no data dir. Only backfillVariableToughness reads it.
var PrintedVariableToughness func(scryfallID string) (top bool, faces []bool, ok bool)

// backfillVariableToughness recomputes Card.VariableToughness (on the
// card, on each of its Faces and on its PrintedSelf) for a restore
// point written before #683 added the field. Such a file has no
// variableToughness key. Decoding that as false is wrong for a `*`
// creature: the first time it lost its last counter, the toughness
// check would put it into the graveyard, while the same card imported
// fresh, or restored by the binary that wrote the file, stays on the
// battlefield.
//
// The source is the Scryfall record for the card's ScryfallID, read
// through PrintedVariableToughness. The importer read that same record
// and applies the same rule, so an imported card comes back flagged
// exactly as a fresh import would flag it. Nothing on a restored Card
// can answer the question by itself. Toughness and Face.Toughness are
// already parsed ints, where a `*` and a printed 0 both became 0, and
// the effect catalog carries no printed stats. A copy carries the
// copied printing's ScryfallID, so a Clone or token copy of a `*`
// creature is answered by the card it copied, and its PrintedSelf by
// its own printing.
//
// A token copy of a double-faced card is the one copy that does not
// line up with its printing: TokenCopyTemplate keeps the ScryfallID
// but drops Faces, so the token has no faces while the printing has
// two or more. Scryfall leaves such a printing's top-level toughness
// empty (Katilda, Dawnhart Martyr; Sage of Ancient Lore), so the top
// answer is always false, and the token copied the flag of whichever
// face was up, which restore cannot see. The flag is set when any
// printed face is variable. That is exact when no face is, and when
// only some are it keeps the skip rather than risk a death a fresh
// token copy of the `*` face would not have.
//
// When the printing is unknown, the flag is set exactly when the
// toughness is 0. That is the skip as it was before #683: the card
// keeps the placeholder exemption after losing its last counter
// instead of gaining a death it never had. The unknown case covers
// every card with no ScryfallID (tokens such as Simulacrum
// Synthesizer's 0/0 Construct, test fixtures, the demo seed), a server
// with no Scryfall dump loaded, and a printing that has left the dump.
// For those cards, and only in games restored from an old file, a
// printed 0/0 that loses its last counter still survives.
//
// A non-zero toughness is never flagged, whatever the printing says:
// the importer only flags the 0 stand-in, and a copy whose exception
// sets a number (Hashaton) clears the flag.
func backfillVariableToughness(c *Card) {
	c.VariableToughness = backfillPrintedVariableToughness(c.ScryfallID, c.Toughness, c.ActiveFace, c.Faces)
	if p := c.PrintedSelf; p != nil {
		p.VariableToughness = backfillPrintedVariableToughness(p.ScryfallID, p.Toughness, p.ActiveFace, p.Faces)
	}
}

// backfillPrintedVariableToughness sets each face's flag in place and
// returns the top-level flag for one set of printed values. With
// faces, the top-level flag is the active face's, because SetFace
// copies it from there.
func backfillPrintedVariableToughness(scryfallID string, toughness, activeFace int, faces []Face) bool {
	var (
		top     bool
		printed []bool
		known   bool
	)
	if PrintedVariableToughness != nil && scryfallID != "" {
		top, printed, known = PrintedVariableToughness(scryfallID)
	}
	facesKnown := known && len(printed) == len(faces)
	for i := range faces {
		faces[i].VariableToughness = faces[i].Toughness == 0 && (!facesKnown || printed[i])
	}
	switch {
	case toughness != 0:
		return false
	case !known:
		return true
	case len(faces) == 0:
		// No faces on the card but two or more on the printing (a
		// token copy of a double-faced card): the copied face is
		// unknown.
		return top || slices.Contains(printed, true)
	case facesKnown && activeFace >= 0 && activeFace < len(faces):
		return printed[activeFace]
	default:
		// The printing's face list does not line up with the card's:
		// treat it as unknown.
		return true
	}
}
