package effects

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spend_only.go — the card-side vocabulary for mana that only some
// spells may use, and costs that only some mana may pay (#1600, ADR
// 0040's 2026-10-03 amendment). The engine half is game/spend_only.go
// (the cost side) and game/mana_restriction.go (the mana side).
//
//	SpendOnlyManaOfTheChosenColor()      "Spend only mana of the chosen color to
//	                                      activate this ability." (Throne of Eldraine)
//	SpendOnlyOnX("B")                    "Spend only black mana on X." (Crypt Rats)
//	MonocoloredSpellsOfTheChosenColor()  "Spend this mana only to cast monocolored
//	                                      spells of that color." (Throne of Eldraine)
//	ManaRestrictMulticolored             "Spend this mana only to cast a multicolored
//	                                      spell." (Pillar of the Paruns)
//	ProducedChosenColorAmount(4)         "Add four mana of the chosen color."
//
// The first two are COST components: compose them into the ability's
// cost with Plus, beside the mana they restrict —
// Plus(ManaCost("{3}"), TapCost(), SpendOnlyManaOfTheChosenColor()).
// The price the client shows stays the printed one; what changes is
// which mana the payment, the auto-tapper, the bot and the view's
// legal actions will accept for it, all through one fold.

// SpendOnlyManaOfTheChosenColor is "Spend only mana of the chosen color
// to activate this ability" — the WHOLE mana cost, paid only with mana
// of the colour the source's "as this enters, choose a color" prompt
// stored (ChooseColorAsEnters). Until a colour is chosen no mana can
// pay it, which is the weaker direction every chosen-colour reader
// takes.
func SpendOnlyManaOfTheChosenColor() game.AbilityCost {
	return game.AbilityCost{SpendOnly: &game.ManaSpendOnly{ChosenColor: true}}
}

// SpendOnlyOnX is "Spend only <colour> mana on X" — the mana paid for
// {X} must be one of `colors` (uppercase WUBRG); the rest of the cost is
// paid as usual. SpendOnlyOnX("B") is Crypt Rats; "black and/or red"
// (Soul Burn) would be SpendOnlyOnX("B", "R").
func SpendOnlyOnX(colors ...string) game.AbilityCost {
	return game.AbilityCost{SpendOnly: &game.ManaSpendOnly{Colors: colors, XOnly: true}}
}

// MonocoloredSpellsOfTheChosenColor is the RestrictionsFunc for "Spend
// this mana only to cast monocolored spells of that color": a cast, of
// a spell that is exactly one colour, and that colour the one stored on
// the source. Before a colour is chosen the colour tag is empty, which
// matches nothing — the source produces no mana then anyway
// (ProducedChosenColorAmount).
func MonocoloredSpellsOfTheChosenColor() func(*game.Game, uuid.UUID, uuid.UUID) []string {
	return func(g *game.Game, _, source uuid.UUID) []string {
		return []string{
			game.ManaRestrictCast,
			game.ManaRestrictMonocolored,
			game.ManaRestrictColor(g.ChosenColorOf(source)),
		}
	}
}

// ProducedChosenColorAmount is the ProducedFunc for "Add N mana of the
// chosen color" — Throne of Eldraine's four. Like ProducedChosenColor
// (its N = 1 case), it produces nothing before a colour is chosen.
func ProducedChosenColorAmount(n int) func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, _, source uuid.UUID) string {
		c := g.ChosenColorOf(source)
		if c == "" || n <= 0 {
			return ""
		}
		return strings.Repeat("{"+c+"}", n)
	}
}

// checkSpendOnlyClause holds a "spend only <colour> mana" clause to the
// shapes the fold can pay (game/spend_only.go), at boot:
//
//   - it restricts a mana component, so the ability must have one;
//   - "on X" needs an {X} to restrict;
//   - it names at least one colour, or the chosen one, and only the
//     five colours — colourless is not a colour (CR 105.1);
//   - not beside a waterbend clause: waterbend's taps pay part of the
//     generic (CR 701.67b), and which part of a restricted cost a tap
//     covers is a question no printed card asks. Refused rather than
//     guessed.
func checkSpendOnlyClause(name string, i int, c game.AbilityCost) {
	s := c.SpendOnly
	if s == nil {
		return
	}
	fail := func(why string) {
		panic(fmt.Sprintf("effects.Register: %q ability %d declares a spend-only clause %s", name, i, why))
	}
	if c.Mana == "" {
		fail("but has no mana component to restrict")
	}
	if s.XOnly && c.XSlots() == 0 {
		fail(fmt.Sprintf("on X but its cost %q has no {X}", c.Mana))
	}
	if !s.ChosenColor && len(s.Colors) == 0 {
		fail("that names no colour — no mana could pay it")
	}
	for _, col := range s.Colors {
		switch col {
		case "W", "U", "B", "R", "G":
		default:
			fail(fmt.Sprintf("naming %q, which is not one of W U B R G", col))
		}
	}
	if c.Waterbend != nil {
		fail("beside a waterbend clause — which part of a restricted cost a tap pays is not modelled")
	}
}
