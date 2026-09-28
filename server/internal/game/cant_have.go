package game

import "fmt"

// cant_have.go — "loses <keyword> and can't have <keyword>" (#1651,
// ADR 0038's amendment of 2026-09-28, B2). Arcane Lighthouse's "until
// end of turn, creatures your opponents control lose hexproof and shroud
// and can't have hexproof or shroud", and the Archetype cycle's
// "creatures your opponents control lose <keyword> and can't have or
// gain <keyword>".
//
// WHY NOT A LAYER-6 REMOVAL. "Can't have" is a CR 101.2 "can't": it
// beats every "can", whatever the timestamps. Layer 6 is applied in
// timestamp order (CR 613.7), so a removal only takes what was granted
// before it — a Heroic Intervention cast afterwards, or a catalog
// creature's own PrintedKeywords static entering afterwards, would put
// the keyword straight back. That was Archetype of Aggression's
// declared gap.
//
// WHY A STRIP AND NOT A FLAG ON THE READER. HasKeyword could answer
// false for a keyword the object can't have, but the keyword would stay
// in Characteristic.Abilities — so the wire would ship a hexproof badge
// on a creature that has none, and every reader that walks the list
// rather than calling HasKeyword (ProtectionQualities, the view) would
// have to learn the rule or leak. Stripping the list once, straight
// after the layer-6 bucket, makes every downstream reader right without
// touching it — materialiseControlLocked's argument, for layer 6.
//
// HOW. A layer-6 effect records the tokens on Characteristic.CantHave
// (cantHaveApply, used by the static form effects.LoseAndCantHave and
// by the scoped mod ModCantHaveKeywords). Recording happens inside the
// ordinary layer-6 Apply, so CR 613.6 needs no code of its own: an
// Archetype that has lost its abilities records nothing. Then
// enforceCantHaveLocked removes every recorded token from Abilities.
// Nothing clears CantHave, a "loses all abilities" included, for the
// reason nothing clears Restrictions: the can't-have is the source's.

// cantHaveApply is the layer-6 Apply that records `keywords` as tokens
// the object can't have. It removes nothing itself: the strip after the
// bucket does, which is what lets it beat a later grant.
func cantHaveApply(keywords []string) func(*Characteristic, *Card, *Game, *Card) {
	return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
		for _, kw := range keywords {
			if !containsFold(ch.CantHave, kw) {
				ch.CantHave = append(ch.CantHave, kw)
			}
		}
	}
}

// CantHaveKeywords is the static form's Apply, for effects.LoseAndCantHave:
// the same recording the scoped mod does, so the two can never disagree.
func CantHaveKeywords(keywords ...string) func(*Characteristic, *Card, *Game, *Card) {
	return cantHaveApply(copyStrings(keywords))
}

// CantHaveKeywordsMod is "loses <keywords> and can't have <keywords>"
// as an ADR 0041 data record (Arcane Lighthouse). Layer 6; reads
// Keywords, which must not be empty.
func CantHaveKeywordsMod(keywords ...string) Mod {
	return Mod{Kind: ModCantHaveKeywords, Keywords: copyStrings(keywords)}
}

// enforceCantHaveLocked strips every token an object can't have from
// its layer-6 ability list. Run once, straight after the layer-6
// bucket (layerPassLocked), so layer 7 and every reader after the pass
// see the stripped list.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) enforceCantHaveLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		ch := g.Battlefield.Cards[i].effective
		if ch == nil || len(ch.CantHave) == 0 || len(ch.Abilities) == 0 {
			continue
		}
		kept := make([]string, 0, len(ch.Abilities))
		for _, a := range ch.Abilities {
			if !containsFold(ch.CantHave, a) {
				kept = append(kept, a)
			}
		}
		ch.Abilities = kept
	}
}

// hexproofModProblem is the registration check for the two #1651
// kinds: a can't-have must name at least one keyword, and a waiver
// names nothing. "" when the mod is sound.
func hexproofModProblem(m Mod) string {
	switch m.Kind {
	case ModCantHaveKeywords:
		if len(m.Keywords) == 0 {
			return "a cantHaveKeywords mod names no keyword"
		}
	case ModWaiveHexproof:
		if len(m.Keywords) != 0 {
			return fmt.Sprintf("a waiveHexproof mod reads no keywords, got %v", m.Keywords)
		}
	}
	return ""
}
