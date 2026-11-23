package game

import "strings"

// flashback_text.go — "a card with flashback or disturb" (Phantom
// Carriage, Shipwreck Sifters, S82 MID batch 7).
//
// A card has flashback when its text prints the keyword (CR 702.34a:
// "Flashback [cost]"). The catalog knows the flashback cost of the cards
// it runs and nothing of the rest; flashback is not a canonical keyword
// token, so the importer's Keywords list drops it; and game.Card carries
// no oracle text. So the deck importer reads the text once and stamps
// Card.FlashbackText, the way it stamps WarpText, and SetFace re-derives
// it from the face that is up. Disturb is a canonical keyword and rides
// Keywords.

// FlashbackInText reports whether any of the texts (the card's and each
// face's) prints flashback as a keyword ability of its own: a line that
// starts "Flashback" and goes on to a cost, "Flashback {2}{R}" or
// "Flashback—Sacrifice three creatures". Reminder text is cut first. A
// sentence that names flashback ("Each instant and sorcery card in your
// graveyard has flashback.") never starts a line with the word, so a
// card that grants flashback or asks about it does not itself read as
// having it.
func FlashbackInText(texts ...string) bool {
	for _, t := range texts {
		for _, line := range strings.Split(t, "\n") {
			if i := strings.IndexByte(line, '('); i >= 0 {
				line = line[:i]
			}
			rest, ok := strings.CutPrefix(strings.ToLower(strings.TrimSpace(line)), "flashback")
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

// HasFlashback reports whether the card has flashback (CR 702.34a),
// whether or not it was ever cast for its flashback cost: the text the
// importer read says so (Card.FlashbackText), or the card's catalog entry
// declares the flashback alternative cost. The catalog answers alone for
// a card the importer never read (PrintedUnknown: a fixture, a restore
// point older than the stamp). A flashback cost another effect granted is
// not printed on the card and does not count.
//
// A face-down card answers false wherever it is (CR 708.2a, CR 406.3a).
func HasFlashback(c Card) bool {
	if c.FaceDown {
		return false
	}
	return c.FlashbackText == PrintedYes || AlternativeCostByKey(CatalogKey(c), "flashback") != nil
}
