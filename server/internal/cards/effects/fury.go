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
// The division is the caster's (#1563, CR 601.2d): the trigger's
// pick_target prompt asks for the targets AND each one's share, every
// target at least 1 and the shares summing to 4. A target that leaves
// in response takes nothing and its share is lost, not moved to the
// others (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:        "fbf9f8c5-849f-45d5-8129-5fc683c21a04",
		Name:            "Fury",
		Completeness:    CompletenessFull,
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
				Or(Creature(), Planeswalker())).WithCount(0, 4).Dividing(Divide(4)),
			Key: "Fury — 4 damage divided among the targets",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDividedDamage(NewContext(g, item))
			},
		}},
	})
}
