package deck

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// rulebreaker_gap_b06.go — S82 GAP batch 6 (#5948): the Rulebreaker commander
// of the Mystery Booster Commander planeswalkers whose deck may hold Aura cards of
// any colour identity and any basic land cards (The Unluckiest Planeswalker). It
// is read off the commander's oracle text, as "can be your commander" is, and
// applied as a pass over the colour-identity violations: a violation for an Aura
// card or a basic land is dropped when a commander of the deck says so.

const gapB06RulebreakerAuras = "can have Aura cards of any color identity and any basic land cards"

// gapB06AurasAnyIdentity reports whether a commander of the deck lets it hold
// Aura cards and basic lands of any colour identity.
func gapB06AurasAnyIdentity(commanders []cards.Card) bool {
	for _, cmd := range commanders {
		texts := []string{cmd.OracleText}
		for _, f := range cmd.CardFaces {
			texts = append(texts, f.OracleText)
		}
		for _, text := range texts {
			if strings.Contains(text, "Rulebreaker") && strings.Contains(text, gapB06RulebreakerAuras) {
				return true
			}
		}
	}
	return false
}

// gapB06DropRulebreakerAuras removes the colour-identity violations of the
// cards a Rulebreaker commander lets into the deck: any Aura card and any basic
// land (never a card that is neither).
func gapB06DropRulebreakerAuras(list *List, vs []Violation) []Violation {
	if !gapB06AurasAnyIdentity(list.Commanders) {
		return vs
	}
	allowed := map[string]bool{}
	for _, c := range list.Mainboard {
		if isBasicLand(c) || strings.Contains(c.TypeLine, "Aura") {
			allowed[c.Name] = true
		}
	}
	out := vs[:0:0]
	for _, v := range vs {
		if v.Code == CodeColorIdentity && allowed[v.Card] {
			continue
		}
		out = append(out, v)
	}
	return out
}
