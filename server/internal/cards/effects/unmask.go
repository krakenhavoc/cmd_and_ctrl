package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unmask — Sorcery {3}{B}:
//
//	"You may exile a black card from your hand rather than pay this
//	 spell's mana cost.
//	 Target player reveals their hand. You choose a nonland card from
//	 it. That player discards that card."
//
// Thoughtseize's pick (ADR 0116) behind Snapback's pitch: the
// alternative cost (CR 118.9) is Pitch with no life, so the exile is
// paid as the spell is cast (CR 601.2h) and a countered Unmask still
// costs the pitched card. Unmask can't pitch itself — it is already on
// the stack when the cost is paid (CR 601.2a).
//
// Targeting yourself reveals your hand to everyone else too (the
// 2014-02-01 ruling); the reveal always goes to the whole table
// (CR 701.20a). A hand with no nonland card is revealed and nothing is
// discarded (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a3a274d-41da-47db-ad41-5720f36a7963",
		Name:         "Unmask",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Pitch("Exile a black card from your hand", 0,
				CardInYourHand("a black card from your hand", OfColor("B")),
				"a black card from your hand"),
		},
		Targets:   TargetPlayer("target player"),
		OnResolve: TargetRevealsYouChooseDiscard(Nonland(), "nonland card"),
	})
}
