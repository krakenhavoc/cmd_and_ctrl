package game

import "strings"

// lost_keywords.go — "a copy of this creature, except it … loses
// <keyword>" (CR 707.9b; Mirage Phalanx's "loses soulbond", #5675).
//
// A copy exception that takes a keyword away is part of the copiable
// values, so it is data on the object (Card.LostKeywords) rather than
// a layer-6 removal: a copy of the copy lacks the keyword too, and an
// effect that later grants the keyword still gives it. The printed
// keyword reaches an object by two roads, the importer's Card.Keywords
// and the catalog's Spec.PrintedKeywords under the copied oracle ID,
// and both are filtered: the first where the exception is applied, the
// second wherever the catalog list is merged (printedInputs,
// PrintedKeywordsHonouringCopy).

// LoseCopiedKeyword is the helper a copy exception calls for "except
// it … loses <keyword>" (Mirage Phalanx's "loses soulbond", #5675,
// CR 707.9b). It strips the keyword from the import road (Keywords) and
// records it (LostKeywords) so the catalog road, where the keyword
// comes from the copied oracle ID's PrintedKeywords, stops merging it
// in. Writing t.Keywords directly in an exception is the bug.
func (c *Card) LoseCopiedKeyword(kw string) {
	c.Keywords, c.LostKeywords = loseCopiedKeyword(c.Keywords, c.LostKeywords, kw)
	c.effective = nil
}

// loseCopiedKeyword is the one body of LoseCopiedKeyword and
// PrintedValues.LoseKeyword: the keyword list without kw, and the lost
// list with it, once.
func loseCopiedKeyword(keywords, lost []string, kw string) ([]string, []string) {
	keywords = withoutKeywordsFold(keywords, []string{kw})
	for _, l := range lost {
		if strings.EqualFold(l, kw) {
			return keywords, lost
		}
	}
	return keywords, append(append([]string(nil), lost...), kw)
}

// withoutKeywordsFold returns list minus every keyword in drop,
// compared without case (the importer stamps Scryfall's "Soulbond",
// the catalog writes "soulbond"). The input is never written through,
// and is returned as it is when nothing is dropped.
func withoutKeywordsFold(list, drop []string) []string {
	if len(drop) == 0 || len(list) == 0 {
		return list
	}
	dropped := func(k string) bool {
		for _, d := range drop {
			if strings.EqualFold(k, d) {
				return true
			}
		}
		return false
	}
	out := list
	for i, k := range list {
		if !dropped(k) {
			continue
		}
		out = append([]string(nil), list[:i]...)
		for _, rest := range list[i+1:] {
			if !dropped(rest) {
				out = append(out, rest)
			}
		}
		break
	}
	return out
}
