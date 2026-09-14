package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reveillark — Creature — Elemental {4}{W}, 4/3 (EDHREC rank 3562):
//
//	"Flying
//	 When this creature leaves the battlefield, return up to two
//	 target creature cards with power 2 or less from your graveyard
//	 to the battlefield.
//	 Evoke {5}{W} (You may cast this spell for its evoke cost. If you
//	 do, it's sacrificed when it enters.)"
//
// The blink deck's value engine. Flying rides PrintedKeywords. The
// leave trigger fires on ANY exit — dying, exile, bounce, a flicker
// (Circuit Mender's shape, not the dies-only gate) — and is targeted
// at "up to two" creature cards in the controller's graveyard with
// power 2 or less, read at announce and again at resolution
// (CR 608.2b); a card that left the graveyard in response is skipped
// and the rest still return. Evoke is the real alternative cost: the
// Lark enters, the sacrifice trigger goes on the stack, and the
// leave trigger fires above the empty board. With no legal card in
// the graveyard the trigger is removed (CR 603.3d) rather than
// prompting.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "1be13ede-98f8-497e-800c-03e5802932b3",
		Name:             "Reveillark",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Evoke("{5}{W}")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCardInGraveyard("up to two target creature cards with power 2 or less from your graveyard",
				YouOwn(), Creature(), PowerLE(2)).WithCount(0, 2),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Reveillark — return up to two small creature cards to the battlefield",
					b34ReturnUpToTwoSmallCreatures)
			},
		}},
	})
}
