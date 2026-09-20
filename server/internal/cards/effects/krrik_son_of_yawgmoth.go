package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// K'rrik, Son of Yawgmoth — Legendary Creature — Phyrexian Horror
// Minion {4}{B/P}{B/P}{B/P}, 2/2:
//
//	"({B/P} can be paid with either {B} or 2 life.)
//	 Lifelink
//	 For each {B} in a cost, you may pay 2 life rather than pay that
//	 mana.
//	 Whenever you cast a black spell, put a +1/+1 counter on K'rrik."
//
// K'rrik's OWN cost needs no declaration. {B/P} is a parsed symbol
// (game.ColorRequirement.Phyrexian) and both payment paths — casting
// a spell and activating an ability (#917 / #971) — already offer the
// life. A six-mana commander that can be cast for {4} and 6 life on
// turn four works out of the box.
//
// # The static that is not here
//
// "For each {B} in a cost, you may pay 2 life rather than pay that
// mana" changes the PAYMENT METHOD of every black pip at the table's
// costs, K'rrik's own and everything else's. The S28 cost-modifier
// engine cannot express it: a CostModifier changes the AMOUNT of
// generic mana, and the Phyrexian flag that unlocks the life option
// is set once by parsing the printed cost, with no hook that lets a
// permanent on the battlefield set it on somebody else's symbol.
//
// Left out, K'rrik is a lifelinking 2/2 that grows on black spells
// and pays for itself with life — weaker than printed, which is the
// only acceptable direction. Written as a fake alternative cost on
// K'rrik alone it would be wrong in both directions at once, so the
// clause waits for a real seam: a payment-method hook beside the
// cost-modifier walk, which is where Phyrexian, convoke and the
// delve family will all eventually read from.
//
// The cast trigger fires once per black spell, checked on the spell's
// COLOUR rather than on a black pip in its cost — a colour-indicator
// or effect-coloured spell counts, and a colourless spell with {B} in
// its cost does not.
func init() {
	Register(Spec{
		OracleID:     "cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd",
		Name:         "K'rrik, Son of Yawgmoth",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"K'rrik's ability to pay 2 life instead of {B} in any cost isn't implemented — only its own printed Phyrexian mana symbols can be paid with life.",
		},
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(OfColor("B"), "K'rrik, Son of Yawgmoth — put a +1/+1 counter on it", putCounterOnSelf),
		},
	})
}
