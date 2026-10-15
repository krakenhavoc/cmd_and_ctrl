package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Charmed Pendant — Artifact {4}:
//
//	"{T}, Mill a card: For each colored mana symbol in the milled
//	 card's mana cost, add one mana of that color. (Activate only as an
//	 instant. For example, if the card's mana cost is {3}{U}{U}{B}, you
//	 add {U}{U}{B}.)"
//
// A mana ability (CR 605.1a: no target, adds mana, not a loyalty
// ability) whose output depends on what its cost milled. The mill is
// MillCardsCost's component (ManaAbilityCost.Mill, #3845), and the
// card it milled is recorded on PaidCost.Milled (#3849, ADR 0020's
// 2026-10-15 amendment), which ProducedForPaid reads after the cost is
// paid. The mana is computed from the card in the graveyard.
//
// Symbols: each {W}, {U}, {B}, {R}, {G} adds that colour; generic,
// colourless ({C}) and {X} symbols add nothing. A Phyrexian symbol
// ({G/P}) or a two-brid ({2/W}) is a symbol of its one colour, so it
// adds that colour. A hybrid symbol of two colours ({W/U}) is a coloured
// symbol of both, and the mana it adds is a pick between them, asked in
// the ordinary mana-colour prompt. A card whose mana cost is empty (a
// land) adds nothing. A card Rest in Peace exiled instead of milling still
// counts: the ability refers to the milled card, and CR 701.17c finds it
// in whatever public zone it went to, so the mana is added.
//
// The auto-tapper never plans the ability (a payment-driven output is
// opaque to it, and milling to pay a spell is the player's call).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12469cb5-a383-47ab-84cd-414b9ced8a8c",
		Name:         "Charmed Pendant",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:            ManaAbilityCost{Tap: true, Mill: 1},
			Label:           "{T}, Mill a card: For each colored mana symbol in the milled card's mana cost, add one mana of that color",
			ProducedForPaid: charmedPendantProduced,
		}},
	})
}

func charmedPendantProduced(g *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	var sb strings.Builder
	for _, id := range paid.Milled {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		sb.WriteString(coloredSymbolsAsProduced(c.ManaCost))
	}
	return game.ManaOutput{Text: sb.String()}
}

// coloredSymbolsAsProduced rewrites a printed mana cost as the mana a
// "for each colored mana symbol" ability adds: one slot per coloured
// symbol, a pipe-separated pick for a symbol of two colours.
func coloredSymbolsAsProduced(cost string) string {
	var out strings.Builder
	for _, raw := range strings.Split(cost, "{") {
		sym := strings.TrimSuffix(raw, "}")
		if sym == "" {
			continue
		}
		var colors []string
		for _, part := range strings.Split(sym, "/") {
			switch part {
			case "W", "U", "B", "R", "G":
				colors = append(colors, part)
			}
		}
		switch len(colors) {
		case 0:
		case 1:
			out.WriteString("{" + colors[0] + "}")
		default:
			out.WriteString("{" + strings.Join(colors, "|") + "}")
		}
	}
	return out.String()
}
