package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fury — Creature — Elemental Incarnation {3}{R}{R}, 3/3 (EDHREC
// rank 2386):
//
//	"Double strike
//	 When this creature enters, it deals 4 damage divided as you
//	 choose among any number of target creatures and/or
//	 planeswalkers.
//	 Evoke—Exile a red card from your hand."
//
// The red incarnation — Solitude's shape with a sweep instead of an
// exile. The evoke is EvokePitch (a red card from hand rather than
// mana, and the CR 702.74a sacrifice trigger bundled with it, so an
// evoked Fury's damage trigger resolves before it dies). The ETB is
// a multi-target clause, one to four targets: "any number" bounded
// above by the four points, since every target must be assigned at
// least one (CR 601.2d).
//
// Declared simplification (weaker than printed): the DIVISION is
// made for the player. The engine's target picker for a triggered
// ability carries no damage distribution — StackItem.Distribution
// exists for a cast, not for the harvester's pick_target prompt — so
// the four points are split as evenly as possible across the chosen
// targets in the order they were picked, the remainder going to the
// earliest picks: one target takes 4, two take 2 each, three take
// 2/1/1, four take 1 each. A 3/1 split is not offered. The seam is
// a distribution on the pick_target prompt.
func init() {
	Register(Spec{
		OracleID:        "fbf9f8c5-849f-45d5-8129-5fc683c21a04",
		Name:            "Fury",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The 4 damage is divided as evenly as possible among the targets you pick, in the order you pick them, rather than however you choose."},
		PrintedKeywords: []string{"double strike"},
		AlternativeCosts: []game.AlternativeCost{
			EvokePitch(
				CardInYourHand("a red card from your hand", OfColor("R")),
				"a red card from your hand",
			),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetPermanent("any number of target creatures and/or planeswalkers",
				Or(Creature(), Planeswalker())).WithCount(0, 4),
			Key: "Fury — 4 damage divided among the targets",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b22DamageDividedEvenly(NewContext(g, item), 4)
			},
		}},
	})
}
