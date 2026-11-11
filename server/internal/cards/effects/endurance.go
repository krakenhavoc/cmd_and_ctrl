package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Endurance — Creature — Elemental Incarnation {1}{G}{G}, 3/4 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"Flash
//	 Reach
//	 When this creature enters, up to one target player puts all the
//	 cards from their graveyard on the bottom of their library in a
//	 random order.
//	 Evoke—Exile a green card from your hand."
//
// The green incarnation: Solitude's evoke (EvokePitch, a green card from
// hand rather than mana, with the CR 702.74a sacrifice trigger bundled)
// over a graveyard reset. "Up to one" lets its controller choose no
// player at all; the chosen player may be its controller. Every card in
// that graveyard as the trigger resolves goes to the bottom of that
// player's library in a random order
// (cmdr2S53TargetPlayerBottomsTheirGraveyard); the library is not
// shuffled. Endurance can't pitch itself: it is already on the stack
// when the cost is paid (CR 601.2a).
//
// Evoked, the enters trigger and the CR 702.74a sacrifice trigger are
// one batch that their controller orders (CR 603.3b), as Fury's are.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c85d824b-c190-4d04-ab99-918ad0e6516c",
		Name:            "Endurance",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
		AlternativeCosts: []game.AlternativeCost{
			EvokePitch(
				CardInYourHand("a green card from your hand", OfColor("G")),
				"a green card from your hand",
			),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Endurance — up to one target player puts their graveyard on the bottom of their library in a random order",
					cmdr2S53TargetPlayerBottomsTheirGraveyard),
				TargetPlayer("up to one target player").WithCount(0, 1),
			),
		},
	})
}
