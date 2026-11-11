package game

import "strings"

// warp_text.go — "target exiled card with warp" (Blade of the Swarm,
// #4461, #5688).
//
// A card in exile has the abilities its text prints (CR 109.3), and a
// card has warp when that text has the keyword (CR 702.185a: "Warp
// [cost]"). The catalog knows the warp cost of the cards it runs and
// nothing of the rest; warp is not a canonical keyword token, so the
// importer's Keywords list drops it; and game.Card carries no oracle
// text. So the deck importer reads the text once and stamps
// Card.WarpText, the way it stamps NoAbilitiesText, and SetFace
// re-derives it from the face that is up.

// WarpInText reports whether any of the texts (the card's and each
// face's) prints warp as a keyword ability of its own: a line that
// starts "Warp" and goes on to a cost, "Warp {2}{R}" or "Warp—[cost]".
// Reminder text is cut first. A sentence that names warp ("Artifact
// cards and red creature cards in your hand have warp {2}{R}.") never
// starts a line with the word, so a card that grants warp or asks about
// it does not itself read as having it.
func WarpInText(texts ...string) bool {
	for _, t := range texts {
		for _, line := range strings.Split(t, "\n") {
			if i := strings.IndexByte(line, '('); i >= 0 {
				line = line[:i]
			}
			rest, ok := strings.CutPrefix(strings.ToLower(strings.TrimSpace(line)), "warp")
			if !ok {
				continue
			}
			rest = strings.TrimSpace(rest)
			if strings.HasPrefix(rest, "{") || strings.HasPrefix(rest, "—") {
				return true
			}
		}
	}
	return false
}

// HasWarp reports whether the card has warp (CR 702.185a), whether or
// not it was ever cast for its warp cost: the text the importer read
// says so (Card.WarpText), or the card's catalog entry declares the
// warp alternative cost. The catalog answers alone for a card the
// importer never read (PrintedUnknown: a fixture, a restore point older
// than the stamp).
//
// A face-down card answers false wherever it is: a face-down permanent
// has no text (CR 708.2a), and a card exiled face down has no
// characteristics (CR 406.3a).
func HasWarp(c Card) bool {
	if c.FaceDown {
		return false
	}
	return c.WarpText == PrintedYes || AlternativeCostByKey(CatalogKey(c), "warp") != nil
}
