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

// RevealYourHand is "Reveal your hand" as a cost (#3250; CR 602.2b,
// 701.20a) — Sasaya, Orochi Ascendant. The whole hand is shown to the
// table as the cost is paid at activation, so there is no count and no
// pick; an empty hand pays it.
//
//	RevealYourHand()
//
// It claims no X and names no card, so it composes with any other
// component (a card file that also spends cards from the hand is
// fine: revealing moves nothing).
func RevealYourHand() game.AbilityCost { return game.AbilityCost{RevealHand: true} }

// checkRevealCardsClause is the boot-time rules for the reveal
// component of an activated ability's cost. `where` names the ability
// in the panic ("ability 0").
func checkRevealCardsClause(name, where string, cost game.AbilityCost) {
	rc := cost.RevealCards
	if rc == nil {
		return
	}
	switch {
	case rc.Behold || rc.Exile:
		panic(fmt.Sprintf("effects.Register: %q %s beholds in an ability cost — behold is a cast's either/or branch (BeholdCost), not an activation component", name, where))
	case rc.Label == "":
		panic(fmt.Sprintf("effects.Register: %q %s reveals cards with no Label — build it with RevealX / RevealN", name, where))
	case rc.Color != "" && !strings.Contains("WUBRG", rc.Color):
		panic(fmt.Sprintf("effects.Register: %q %s reveals cards of colour %q — one of W, U, B, R, G", name, where, rc.Color))
	case rc.AnyNumber:
		panic(fmt.Sprintf("effects.Register: %q %s reveals any number of cards in an activation cost — that form is a mana ability's (ManaAbilityCost.Reveal, #6085)", name, where))
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

// RevealAnyNumber is a mana ability's "Reveal any number of <quality>
// cards from your hand" (#6085): Rosheen, Roaring Prophet's "{T}: Reveal
// any number of cards with {X} in their mana cost in your hand. Add
// {C}{C} for each card revealed this way." Put it on
// ManaAbilityCost.Reveal; the activator names the cards on `reveal_ids`
// (zero is a legal answer), they are revealed as the ability resolves,
// and ProducedForPaid reads them back as PaidCost.Revealed. `quality` is
// the RevealCost the cards must match; the label is the clause as
// printed, without the verb.
func RevealAnyNumber(label string, quality game.RevealCost) *game.RevealCardsCost {
	return &game.RevealCardsCost{RevealCost: quality, AnyNumber: true, Label: label}
}

// checkManaRevealClause is the boot-time rule for a mana ability's
// reveal (#6085): the AnyNumber form only, with a label, a hand quality
// (no behold, no exile) and something that reads the revealed cards.
func checkManaRevealClause(name string, i int, ma ManaAbility) {
	rc := ma.Cost.Reveal
	if rc == nil {
		return
	}
	switch {
	case !rc.AnyNumber || rc.N != 0 || rc.CountFromX || rc.SharesColor:
		panic(fmt.Sprintf("effects.Register: %q mana ability %d reveals a fixed or X count — a mana ability announces no X (CR 605.3b); build it with RevealAnyNumber", name, i))
	case rc.Label == "":
		panic(fmt.Sprintf("effects.Register: %q mana ability %d reveals cards with no Label — build it with RevealAnyNumber", name, i))
	case rc.Behold || rc.Exile || rc.Chosen || rc.ChosenType || rc.ChooseCreatureOrWarped:
		panic(fmt.Sprintf("effects.Register: %q mana ability %d reveals with a cast-only RevealCost form", name, i))
	case rc.Color != "" && !strings.Contains("WUBRG", rc.Color):
		panic(fmt.Sprintf("effects.Register: %q mana ability %d reveals cards of colour %q — one of W, U, B, R, G", name, i, rc.Color))
	case ma.ProducedForPaid == nil:
		panic(fmt.Sprintf("effects.Register: %q mana ability %d reveals cards but nothing reads them — declare ProducedForPaid", name, i))
	}
}
