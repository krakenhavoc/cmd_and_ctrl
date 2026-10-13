package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flash_if_x_at_most.go — "You may cast this spell as though it had flash
// if X is N or less" (Silver Scrutiny; #3225).
//
// CR 601.3c: an effect that lets a spell be cast as though it had flash
// only if a condition holds lets its caster BEGIN to cast it at instant
// speed; the condition is then checked against the announcement. Here
// the condition is on X, a number the caster chooses after the timing
// window has been opened, so the shape is the same as FlashForMore's —
// an alternative cost that opens the window for that claim only
// (AlternativeCost.AsThoughFlash) — with the bound carried on the offer
// (AlternativeCost.XCeiling) rather than on the card.
//
// It has to be the offer's, not the card's. Spec.XCeiling would bound
// every cast of the spell and make a sorcery-speed Silver Scrutiny for
// X = 10 illegal, which nothing prints; an unbounded flash offer would
// let a player draw ten cards at instant speed, which is stronger than
// printed. The offer charges the card's own mana cost, so claiming it
// costs nothing extra and the price is the same as the printed one.

// FlashIfXAtMost is the offer. `cost` is the card's printed mana cost
// (it must contain {X}, which Register checks) and `max` is the largest
// X the flash window allows.
//
//	AlternativeCosts: []game.AlternativeCost{FlashIfXAtMost("{X}{U}{U}", 3)},
func FlashIfXAtMost(cost string, max int) game.AlternativeCost {
	return game.AlternativeCost{
		Key:           "flash-x-at-most",
		Label:         fmt.Sprintf("Cast with flash — X is %d or less (%s)", max, cost),
		ManaCost:      cost,
		AsThoughFlash: true,
		XCeiling:      max,
	}
}

// FlashIfTargets is "You may cast this spell as though it had flash if it
// targets <what>" (Timely Ward: a commander). Same shape as FlashIfXAtMost:
// an offer at the card's printed cost that opens the instant-speed window
// for its own claim, here carrying the NARROWER target clause the flash
// cast must satisfy (AlternativeCost.Targets). The printed clause still
// governs the sorcery-speed cast; claiming the offer is what narrows it,
// so the spell cannot be flashed onto a creature that isn't a commander.
//
//	AlternativeCosts: []game.AlternativeCost{FlashIfTargets("{2}{W}", "a commander", EnchantCreature(ACommander()))}
func FlashIfTargets(cost, what string, clause *game.TargetSpec) game.AlternativeCost {
	return game.AlternativeCost{
		Key:           "flash-if-targets",
		Label:         fmt.Sprintf("Cast with flash — it targets %s (%s)", what, cost),
		ManaCost:      cost,
		AsThoughFlash: true,
		Targets:       clause,
	}
}
