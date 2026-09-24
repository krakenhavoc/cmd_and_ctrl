package game

// coverage.go answers one question, and it is the player's question
// rather than the engine's: "will this card do what it says?"
//
// The effect catalog is opt-in (ADR 0010). A few hundred cards have a
// hand-written Spec; the other ~31,500 Commander-legal ones resolve
// with no effect at all and the players move the pieces by hand,
// Cockatrice-style. That is a deliberate design decision and it is
// not going to change soon — the catalog grows one card file at a
// time. The defect is not the gap, it is the SILENCE about the gap:
// five of the twelve reports from the 2026-09-10 playtest (#321,
// #324, #325, #332, #333) are one player casting five different
// uncatalogued cards, watching nothing happen, and reasonably filing
// five bugs.
//
// The naive signal — "this card is not in the catalog" — would be a
// lie, and a loud one, because catalog membership is not the same
// thing as working. Since #317 / #319 / #320 the engine honours
// printed keywords on every card in the dump, so an uncatalogued
// Serra Angel flies, has vigilance, blocks correctly and deals its
// damage; an uncatalogued Grizzly Bears is a complete and correct
// 2/2; a Forest taps for {G} off its type line alone (CR 305.6).
// Badging those would be crying wolf on most of a real deck.
//
// The honest predicate is narrower: a card is unimplemented when it
// PRINTS RULES THE ENGINE WILL NOT RUN. That is two independent
// facts, and they are joined in exactly one place — Unimplemented —
// so what the player is told at deck import is the same thing they
// are told at cast:
//
//	NeedsCatalogEffect  a pure function of the printed Scryfall
//	                    record, stamped onto Card.NeedsEffect by the
//	                    deck importer alongside Keywords and
//	                    StartingLoyalty, on the road #274 built.
//	IsAutoCard          is there a Spec? Already the `auto` bit.

import "strings"

// NeedsCatalogEffect reports whether a card's printed characteristics
// describe rules that only a hand-written catalog Spec can carry out.
//
// False means the engine is already complete for this card without
// one: a vanilla creature, a creature whose entire text is keyword
// abilities the engine enforces, a basic land. True means the card
// prints something the engine has no generic path for, so without a
// Spec it will sit on the battlefield doing nothing.
//
// typeLine is Scryfall's type line. texts is the top-level
// oracle_text followed by each face's oracle_text — a multi-faced
// card leaves the top level empty and puts its text on the faces, so
// callers pass the union. That deliberately over-counts a
// double-faced card with a vanilla front face, which is the right
// answer anyway: multi-face cards are not modelled at all (ADR 0034),
// so the engine will not do what such a card says regardless of which
// half prints the text.
func NeedsCatalogEffect(typeLine string, texts ...string) bool {
	// A basic land's mana ability comes from its type line (CR
	// 305.6) and is printed only as reminder text, so the text scan
	// below can never see it. ManaAbilitiesForCard synthesises that
	// ability for the five basic land types and for nothing else,
	// which leaves Wastes: all reminder text, no colour, and no
	// mana at all without a Spec. Two cards in the format hit this,
	// and getting them right is the difference between a predicate
	// that is honest about lands and one that merely looks like it.
	if typeLineHas(typeLine, "basic") && typeLineHas(typeLine, "land") &&
		len(intrinsicLandManaAbilities(Card{TypeLine: typeLine})) == 0 {
		return true
	}
	for _, text := range texts {
		if printsUnhandledRules(text) {
			return true
		}
	}
	return false
}

// Unimplemented reports whether the engine will fail to carry out
// this card's printed rules: it needs a catalog Spec and does not
// have one.
//
// Cards that never went through deck import — tokens, test fixtures,
// the demo seed — carry NeedsEffect == false and so are never
// flagged. That is the safe default on purpose: the cost of a missed
// signal is a player who learns what they would have learned anyway,
// and the cost of a false one is a signal nobody trusts.
//
// A nil IsCatalogCard hook (a binary built without the effects blank
// import) makes every card look uncatalogued. That is not a bug in
// this predicate — with no catalog wired, nothing auto-resolves, and
// every card really is unimplemented in that binary.
func Unimplemented(c Card) bool {
	// CR 708.2a: a face-down permanent has no printed rules for the
	// engine to fail to carry out, so it is never flagged. The guard
	// is explicit rather than left to CatalogKey's empty key, which
	// would make the answer `c.NeedsEffect` — flagging every
	// face-down permanent that happens to be a card with text, and
	// telling the table something about a card it cannot read.
	if c.FaceDownIsPermanent() {
		return false
	}
	// #522: a restore brought this card back with fewer catalog
	// abilities than were captured, so the engine is no longer
	// running the rules the table saw it run. Flagged for the rest of
	// the game, whatever the card prints and whatever is left of its
	// entry — see Card.AbilitiesLostOnRestore.
	if c.AbilitiesLostOnRestore {
		return true
	}
	return c.NeedsEffect && !IsAutoCard(CatalogKey(c))
}

// UnimplementedNames returns the distinct names of the cards in the
// list whose printed rules the engine will not carry out, in printed
// order with duplicates collapsed — a deck runs four Lightning Bolts
// and wants to be told about Lightning Bolt once. Used by the deck
// importer to tell a player what they are signing up for before the
// game starts, which is a far better moment to find out than the
// middle of a combat.
func UnimplementedNames(cards []Card) []string {
	seen := make(map[string]bool, len(cards))
	out := make([]string, 0, 8)
	for _, c := range cards {
		if !Unimplemented(c) || seen[c.Name] {
			continue
		}
		seen[c.Name] = true
		out = append(out, c.Name)
	}
	return out
}

// printsUnhandledRules reports whether one face's oracle text has a
// line that is not a keyword-ability line.
//
// Oracle text is newline-separated abilities. A keyword-ability line
// holds nothing but keywords, comma-separated — "Flash", "Flying,
// vigilance", "Reach, trample". Anything else standing on its own
// line is a rule: a triggered ability, an activated ability, a static
// effect, a spell instruction. Reminder text is removed first; it is
// parenthesised by definition and never carries rules, so a Forest's
// "({T}: Add {G}.)" and a flier's "(This creature can't be blocked
// except by creatures with flying or reach.)" both vanish before the
// scan.
//
// Keywords the engine does NOT enforce — ward, cycling, equip,
// flashback, the whole long tail — read as rules by this test, and
// that is the correct answer rather than a rough edge.
// canonicalKeywords is a closed set precisely because a keyword joins
// it in the same change that teaches the engine to honour it
// (keywords.go), so a keyword outside the table is by construction a
// keyword nothing acts on.
//
// #662: the same holds one level down for PROTECTION, whose quality
// is a parameter. "Protection from red" is a keyword line the engine
// honours; "protection from monocolored" is a rule it does not, and
// the closed grammar in protection.go is what tells them apart. Both
// answers come from CanonicalKeywords, so the coverage badge and the
// deck importer can never disagree about which is which.
func printsUnhandledRules(text string) bool {
	for _, line := range strings.Split(stripReminderText(text), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !keywordOnlyLine(line) {
			return true
		}
	}
	return false
}

// keywordOnlyLine reports whether every comma-separated item on the
// line is a keyword the engine enforces. Empty items are skipped so a
// stray trailing comma doesn't sink an otherwise clean keyword line;
// a line that is nothing but punctuation therefore finds no keyword
// at all and returns false, which the caller reads as "a rule" — the
// conservative direction for text this function does not understand.
func keywordOnlyLine(line string) bool {
	found := false
	for _, part := range strings.Split(line, ",") {
		if strings.TrimSpace(part) == "" {
			continue
		}
		// CanonicalKeywords, not the singular form: one printed
		// clause can be two abilities ("protection from Demons and
		// from Dragons", CR 702.16m).
		if _, ok := CanonicalKeywords(part); !ok {
			return false
		}
		found = true
	}
	return found
}

// stripReminderText removes every parenthesised span. Scryfall's
// reminder text is always parenthesised and is never nested, but the
// depth counter costs nothing and makes the function total over
// arbitrary input. An unclosed "(" swallows the rest of the string,
// newlines included, which errs toward NOT flagging the card — the
// direction this whole file errs in.
func stripReminderText(text string) string {
	if !strings.ContainsRune(text, '(') {
		return text
	}
	var b strings.Builder
	b.Grow(len(text))
	depth := 0
	for _, r := range text {
		switch {
		case r == '(':
			depth++
		case r == ')' && depth > 0:
			depth--
		case depth == 0:
			b.WriteRune(r)
		}
	}
	return b.String()
}
