package game

// split_second.go — CR 702.61, read off the card (#1519).
//
// The rule was built in S13.1 and never switched on. Every consumer
// already existed and already asked one cache, Game.SplitSecondActive:
// CastSpell, both activation paths and ActivateLoyalty refuse with
// ErrSplitSecondActive, the enumerator's castMoves and activatedMoves
// return early, suspend's special-action window shuts, and the wire
// carries `split_second_active`. What was missing was the WRITER: the
// only thing that ever stamped StackItem.SplitSecond was the sandbox
// `split_second` flag on cast_spell, which no client sends. So a
// Krosan Grip on the stack could be answered like any other spell.
//
// The keyword now joins canonicalKeywords, so the deck importer stops
// filtering Scryfall's "Split second" out of Card.Keywords, and the
// cast path stamps the stack item from the spell's own keywords
// through castHasSplitSecond below. The sandbox flag is routed through
// the same function rather than beside it, so there is one writer and
// one answer.
//
// What split second does NOT stop (CR 702.61b), and the engine
// already honoured before any card could turn it on:
//
//   - mana abilities (ActivateManaAbility has no split-second check,
//     and ActivationTimingOpenLocked returns before its own);
//   - special actions — foretell, turning a face-down permanent face
//     up (SpecialActionTimingOKLocked; suspend is the exception,
//     CR 702.62c, because its window is the card's casting window);
//   - triggered abilities, which trigger and go on the stack as
//     usual. A trigger whose resolution would CAST a spell cannot
//     (Gatherer ruling on Sudden Shock, 2006-09-25), which CastSpell's
//     own check answers for free.
//
// The cache is refreshed by recomputeSplitSecondLocked after every
// stack exit, so it ends the moment the spell resolves, is countered,
// fizzles or is exiled off the stack. See ADR 0007's amendment of
// 2026-09-24.

// KeywordSplitSecond is CR 702.61's canonical token — Scryfall's
// "Split second", lowercased.
const KeywordSplitSecond = "split second"

// castHasSplitSecond reports whether a spell being cast has split
// second, which is the fact StackItem.SplitSecond records. `card` is
// CastSpell's value copy with the chosen face already materialised,
// so a multi-face card answers for the face being cast.
//
// Two sources, one answer:
//
//   - the card's own keywords, read through HasKeyword: Card.Keywords
//     for a deck-imported card and CatalogPrintedKeywords for a
//     catalog entry that declares it (which is how a fixture with no
//     Scryfall data gets it). Off the battlefield there is no layer
//     engine, so no effect can grant or remove it — no card in print
//     grants split second to a spell anyway;
//   - the sandbox `split_second` flag on cast_spell (S13.1), kept so
//     a table playing a card the catalog has never heard of can still
//     say so by hand.
//
// A spell cast FACE DOWN has neither: CR 708.2 makes it a 2/2 with no
// text, and CR 708.4 is explicit that the face-down spell is that
// object while it is on the stack. The flag is refused too, because a
// face-down spell announcing split second would tell the table which
// card it is.
func castHasSplitSecond(card *Card, faceDown FaceDownKind, sandboxFlag bool) bool {
	if faceDown != FaceDownNone {
		return false
	}
	return sandboxFlag || HasKeyword(card, KeywordSplitSecond)
}
