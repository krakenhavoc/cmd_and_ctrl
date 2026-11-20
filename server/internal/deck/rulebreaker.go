package deck

import (
	"fmt"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// rulebreaker.go — the "Rulebreaker" commanders of Mystery Booster Commander
// (S82 GAP batch 5, #5948): a commander whose own text changes what its deck may
// hold. Read off the oracle text, as "can be your commander" is, so the table
// of commanders is the card pool and not a list here.
//
//   - Seluma, Light of Aysen: "A deck with this commander can have Angel cards
//     of any color identity and any basic land cards."
//   - Whtz, the Bibliophile: "A deck with this commander has no maximum deck
//     size." The deck still holds at least 100 cards (CR 903.5a).
//   - Tolabow, Loch Rascal: "the color identity of instant and sorcery cards in
//     your deck can include one color of your choice not in your commander's
//     color identity, and your deck can have any basic land cards."
//
// The player's "color of your choice" is not a field of a deck list: a deck
// qualifies when one color outside the commander's identity covers every
// off-identity instant and sorcery in it.

// rulebreakerClauses are the oracle phrases this file understands.
const (
	rulebreakerAngels = "can have Angel cards of any color identity and any basic land cards"
	rulebreakerNoMax  = "A deck with this commander has no maximum deck size"
	rulebreakerSpells = "the color identity of instant and sorcery cards in your deck can include one color of your choice not in your commander's color identity"
)

// rulebreaker is what a deck's commanders let it hold.
type rulebreaker struct {
	angelsAnyIdentity bool
	basicsAnyIdentity bool
	oneSpellColor     bool
	noMaxDeckSize     bool
}

func commanderRulebreaker(commanders []cards.Card) rulebreaker {
	var rb rulebreaker
	for _, cmd := range commanders {
		texts := []string{cmd.OracleText}
		for _, f := range cmd.CardFaces {
			texts = append(texts, f.OracleText)
		}
		for _, text := range texts {
			if !strings.Contains(text, "Rulebreaker") {
				continue
			}
			if strings.Contains(text, rulebreakerAngels) {
				rb.angelsAnyIdentity, rb.basicsAnyIdentity = true, true
			}
			if strings.Contains(text, rulebreakerNoMax) {
				rb.noMaxDeckSize = true
			}
			if strings.Contains(text, rulebreakerSpells) {
				rb.oneSpellColor, rb.basicsAnyIdentity = true, true
			}
		}
	}
	return rb
}

func isAngelCard(c cards.Card) bool { return strings.Contains(c.TypeLine, "Angel") }

func isInstantOrSorcery(c cards.Card) bool {
	return strings.Contains(c.TypeLine, "Instant") || strings.Contains(c.TypeLine, "Sorcery")
}

// offIdentity is the symbols of c's color identity the commander's identity
// does not hold, in the card's order.
func offIdentity(c cards.Card, allowed map[string]struct{}) []string {
	var off []string
	for _, sym := range c.ColorIdentity {
		if _, ok := allowed[sym]; !ok {
			off = append(off, sym)
		}
	}
	return off
}

// colorIdentityViolations is the commander colour-identity check with the
// commanders' Rulebreaker clauses applied. A card with no clause in play is
// judged as it always was: one violation for any symbol outside the identity.
func colorIdentityViolations(list *List, allowed map[string]struct{}) []Violation {
	rb := commanderRulebreaker(list.Commanders)
	allowedList := commanderIdentityList(allowed)
	var vs []Violation
	spellColor := map[string][]cards.Card{} // the one off-identity colour of an instant or sorcery
	for _, c := range list.Mainboard {
		off := offIdentity(c, allowed)
		if len(off) == 0 {
			continue
		}
		switch {
		case rb.basicsAnyIdentity && isBasicLand(c):
			continue
		case rb.angelsAnyIdentity && isAngelCard(c):
			continue
		case rb.oneSpellColor && isInstantOrSorcery(c) && len(off) == 1:
			spellColor[off[0]] = append(spellColor[off[0]], c)
			continue
		}
		vs = append(vs, Violation{
			Code:    CodeColorIdentity,
			Card:    c.Name,
			Message: fmt.Sprintf("%q has color identity %v but commander allows %v", c.Name, c.ColorIdentity, allowedList),
		})
	}
	// One colour for every such spell: keep the colour with the most cards
	// (ties go to the earlier one in WUBRG order) and flag the rest.
	if len(spellColor) > 1 {
		keep, most := "", 0
		syms := make([]string, 0, len(spellColor))
		for sym := range spellColor {
			syms = append(syms, sym)
		}
		sort.Slice(syms, func(i, j int) bool { return strings.Index("WUBRG", syms[i]) < strings.Index("WUBRG", syms[j]) })
		for _, sym := range syms {
			if len(spellColor[sym]) > most {
				keep, most = sym, len(spellColor[sym])
			}
		}
		for _, sym := range syms {
			if sym == keep {
				continue
			}
			for _, c := range spellColor[sym] {
				vs = append(vs, Violation{
					Code: CodeColorIdentity,
					Card: c.Name,
					Message: fmt.Sprintf("%q has color identity %v: instants and sorceries may include only one color outside the commander's (%s here)",
						c.Name, c.ColorIdentity, keep),
				})
			}
		}
	}
	return vs
}
