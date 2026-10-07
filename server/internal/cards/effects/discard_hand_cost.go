package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_hand_cost.go — #1600: the boot-time rules for a discard
// cost component, shared by its two owners (an activated ability's
// AbilityCost.DiscardCards, a mana ability's ManaAbilityCost.
// DiscardCards) so they cannot drift.
//
// Two forms (game.DiscardCost):
//
//   - a COUNT ("Discard a card", "Discard two cards", "… at random"):
//     at least one card, because a clause that discards nothing would
//     make the ability free (#660, #1213);
//   - "Discard your hand" (DiscardCost.Hand, built by DiscardYourHand):
//     no count at all — the hand is the count, and an empty hand pays
//     it (CR 118.3). A count, a predicate or "at random" beside it is a
//     card file that has mixed the two forms up.
//
// And one rule across components: a hand clause empties the hand as
// the cost is paid, so another component that spends a card FROM the
// hand — cycling's "Discard this card", "Exile a card from your hand",
// "Put a card from your hand on top of your library" — could never be
// paid beside it (CR 118.3: one card pays one cost). No printed card
// has the combination; a card file that declared it would register an
// ability nobody can activate.

// checkDiscardClause is the per-clause half: the count form demands a
// card, the hand form demands no count. `where` names the ability in
// the panic ("ability 0", "mana ability 1").
func checkDiscardClause(name, where string, dc *game.DiscardCost) {
	if dc == nil {
		return
	}
	// #2527: "Discard X cards" counts from the announced X, so N is
	// never read; a fixed count, "at random" or the hand form beside it
	// is a card file that has mixed the forms up. An ABILITY only is
	// checked by checkDiscardXClause, which knows the owner.
	if dc.CountFromX {
		if dc.N != 0 || dc.Random || dc.Hand {
			panic(fmt.Sprintf("effects.Register: %q %s discards X cards and also declares a count, \"at random\" or your hand — build it with DiscardX",
				name, where))
		}
		return
	}
	if dc.Hand {
		if dc.N != 0 || dc.Match != nil || dc.Random {
			panic(fmt.Sprintf("effects.Register: %q %s discards your hand and also declares a count, a predicate or \"at random\" — build it with DiscardYourHand",
				name, where))
		}
		return
	}
	// #660 / #1213: a count clause that discards nothing would make the
	// ability free, the way a zero-counter cost would.
	if dc.N <= 0 {
		panic(fmt.Sprintf("effects.Register: %q %s discards %d cards — a discard cost discards at least one",
			name, where, dc.N))
	}
}

// checkDiscardXBesideOtherCosts is the cross-component half of the
// "Discard X cards" form (#2527), for an activated ability's cost:
//
//   - one announced X cannot pay two clauses, so {X} in the mana
//     component, a sacrifice-X clause or a tap-X clause beside it is
//     refused (the rule SacrificeX and TapXUntapped already follow);
//   - the cards are named in `discard_ids` out of the hand, and the
//     other hand-spending components (discard this, put a card from
//     the hand on top, exile a card from the hand) name theirs out of
//     the same hand with a count the ladder of X cannot reserve. No
//     printed cost combines them, and the enumerator's X ladder
//     assumes it, so a card file that does is refused rather than
//     offered moves the engine bounces.
func checkDiscardXBesideOtherCosts(name, where string, cost game.AbilityCost) {
	if !game.DiscardCountFromX(cost.DiscardCards) {
		return
	}
	if cost.EnergyX {
		panic(fmt.Sprintf("effects.Register: %q %s pays X energy AND discards X cards — one announced X cannot pay both",
			name, where))
	}
	if cost.XSlots() > 0 {
		panic(fmt.Sprintf("effects.Register: %q %s has {X} in its mana cost %q AND discards X cards — one announced X cannot pay both",
			name, where, cost.Mana))
	}
	if game.SacrificeCountFromX(cost.SacrificeOther) || game.TapOthersCountFromX(cost.TapOthers) {
		panic(fmt.Sprintf("effects.Register: %q %s discards X cards and also sacrifices or taps X permanents — one announced X cannot pay both",
			name, where))
	}
	if cost.DiscardSelf || cost.PutFromHandOnLibraryTop > 0 ||
		(cost.ExileCards != nil && cost.ExileCards.Zone() == game.ZoneHand) {
		panic(fmt.Sprintf("effects.Register: %q %s discards X cards and also spends another card from the hand — no printed cost does both",
			name, where))
	}
}

// checkDiscardHandBesideHandCosts is the cross-component half, for an
// activated ability's cost.
func checkDiscardHandBesideHandCosts(name, where string, cost game.AbilityCost) {
	if !cost.DiscardCards.DiscardsHand() {
		return
	}
	if cost.DiscardSelf || cost.PutFromHandOnLibraryTop > 0 ||
		(cost.ExileCards != nil && cost.ExileCards.Zone() == game.ZoneHand) {
		panic(fmt.Sprintf("effects.Register: %q %s discards your hand and also spends a card from it — no printed cost does both, and it could never be paid",
			name, where))
	}
}

// checkManaDiscardHandBesideHandCosts is the same rule for a mana
// ability's cost, whose hand-card components are the exile-a-card and
// exile-this-card ones (a Spirit Guide's source is in the hand).
func checkManaDiscardHandBesideHandCosts(name, where string, cost ManaAbilityCost) {
	if !cost.DiscardCards.DiscardsHand() {
		return
	}
	if cost.ExileSelf || (cost.ExileCards != nil && cost.ExileCards.Zone() == game.ZoneHand) {
		panic(fmt.Sprintf("effects.Register: %q %s discards your hand and also spends a card from it — no printed cost does both, and it could never be paid",
			name, where))
	}
}
