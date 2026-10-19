package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Charmed Pendant — Artifact {4}:
//
//	"{T}, Mill a card: For each colored mana symbol in the milled
//	 card's mana cost, add one mana of that color. (Activate only as an
//	 instant. For example, if the card's mana cost is {3}{U}{U}{B}, you
//	 add {U}{U}{B}.)"
//
// An ordinary activated ability whose output depends on what its cost
// milled. It is not a mana ability (#3971): the cost mills a card from a
// library, which CR 605.1a excludes, so it uses the stack and the mana
// is added on resolution. The mill is MillCardsCost (AbilityCost.Mill,
// #3845), and the card it milled is recorded on PaidCost.Milled (#3849,
// ADR 0020's 2026-10-15 amendment), which the resolving item carries
// (Context.Paid). The mana is computed from that card when the ability
// resolves, wherever it now is.
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
// The auto-tapper never plans the ability (milling to pay a spell is the
// player's call).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12469cb5-a383-47ab-84cd-414b9ced8a8c",
		Name:         "Charmed Pendant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Mill a card: For each colored mana symbol in the milled card's mana cost, add one mana of that color",
			Cost:    Plus(TapCost(), MillCardsCost(1)),
			Purpose: game.Purpose{Answers: game.AnswerValue}, // ADR 0142: mana answers nothing
			Effect:  charmedPendantAddMana,
		}},
	})
}

// charmedPendantAddMana adds the mana the milled card's colored symbols
// name, to the ability's controller, as it resolves.
func charmedPendantAddMana(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var sb strings.Builder
	for _, id := range ctx.Paid().Milled {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		sb.WriteString(coloredSymbolsAsProduced(c.ManaCost))
	}
	return AddMana{Produced: sb.String()}.Apply(ctx)
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
