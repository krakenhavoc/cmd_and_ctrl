package effects

import (
	"fmt"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_cards_cost.go — #2598: "Reveal N <quality> cards from your
// hand" as the cost of an activated ability (ADR 0020's 2026-10-08
// amendment). The Martyr cycle's "{1}, Reveal X black cards from your
// hand, Sacrifice this creature:".

// RevealX is "Reveal X <colour> cards from your hand" as a cost. The
// count is the X announced with the activation (CR 602.2b), which the
// effect reads back with ctx.X(); the activator names that many cards
// in hand (`reveal_ids`). There is no {X} symbol, so the mana cost is
// the printed one whatever X is, and X may be zero.
//
//	Plus(ManaCost("{1}"), RevealX("X black cards", "B"), SacrificeThis())
//
// `color` is one of "W", "U", "B", "R", "G" (empty is "any card"). The
// label is the clause as printed, without the verb. Register refuses
// it beside another claim on the announced X (an {X} in the mana cost,
// a sacrifice-X, tap-X or discard-X clause, pay-X-energy) — one number
// cannot pay two clauses — and it cannot be put on a mana ability, so
// it never needs a stack item to carry its X (CR 605.3b).
func RevealX(label, color string) game.AbilityCost {
	return game.AbilityCost{RevealCards: &game.RevealCardsCost{
		RevealCost: game.RevealCost{Color: color},
		CountFromX: true,
		Label:      label,
	}}
}

// RevealN is "Reveal N <colour> cards from your hand" with a fixed
// count. No printed card in the catalog uses it yet; it is the same
// component with a printed number, kept so the first one is a card
// file and not an engine change.
func RevealN(n int, label, color string) game.AbilityCost {
	return game.AbilityCost{RevealCards: &game.RevealCardsCost{
		RevealCost: game.RevealCost{Color: color},
		N:          n,
		Label:      label,
	}}
}

// checkRevealCardsClause is the boot-time rules for the reveal
// component of an activated ability's cost. `where` names the ability
// in the panic ("ability 0").
func checkRevealCardsClause(name, where string, cost game.AbilityCost) {
	rc := cost.RevealCards
	if rc == nil {
		return
	}
	switch {
	case rc.Behold:
		panic(fmt.Sprintf("effects.Register: %q %s beholds in an ability cost — behold is a cast's either/or branch (BeholdCost), not an activation component", name, where))
	case rc.Label == "":
		panic(fmt.Sprintf("effects.Register: %q %s reveals cards with no Label — build it with RevealX / RevealN", name, where))
	case rc.Color != "" && !strings.Contains("WUBRG", rc.Color):
		panic(fmt.Sprintf("effects.Register: %q %s reveals cards of colour %q — one of W, U, B, R, G", name, where, rc.Color))
	case rc.CountFromX && rc.N != 0:
		panic(fmt.Sprintf("effects.Register: %q %s reveals X cards and also declares a fixed count — build it with RevealX", name, where))
	case !rc.CountFromX && rc.N <= 0:
		panic(fmt.Sprintf("effects.Register: %q %s reveals %d cards — a reveal cost reveals at least one, or X", name, where, rc.N))
	}
	if !rc.CountFromX {
		return
	}
	// One announced X cannot pay two clauses (the rule every other
	// X-count component follows).
	switch {
	case cost.XSlots() > 0:
		panic(fmt.Sprintf("effects.Register: %q %s has {X} in its mana cost %q AND reveals X cards — one announced X cannot pay both", name, where, cost.Mana))
	case cost.EnergyX:
		panic(fmt.Sprintf("effects.Register: %q %s pays X energy AND reveals X cards — one announced X cannot pay both", name, where))
	case game.SacrificeCountFromX(cost.SacrificeOther) || game.TapOthersCountFromX(cost.TapOthers):
		panic(fmt.Sprintf("effects.Register: %q %s reveals X cards and also sacrifices or taps X permanents — one announced X cannot pay both", name, where))
	case game.DiscardCountFromX(cost.DiscardCards) || game.DiscardManaValueX(cost.DiscardCards):
		panic(fmt.Sprintf("effects.Register: %q %s reveals X cards and also discards X cards — one announced X cannot pay both", name, where))
	}
}
