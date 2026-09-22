package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forum Familiar — 1/1 Cat for {W}:
//
//	"Disguise {1}{W}. When this creature is turned face up, return
//	 another target permanent you control to its owner's hand and put
//	 a +1/+1 counter on this creature."
//
// The disguise card in the catalog, and the one that exercises the
// whole of #1194 end to end rather than a corner of it:
//
//   - cast face down for {3} it is a nameless 2/2 — with WARD {2}
//     (CR 702.168a), which is the one ability a CR 708.2 object has and
//     the only thing separating disguise from morph. It is not a
//     keyword on the card; it comes from the face-down STATE, so it is
//     there for as long as the permanent is face down and gone the
//     instant it is not (ADR 0082 decision 8).
//   - {1}{W} at any time turns it face up (CR 708.6), and
//   - the CR 708.8 trigger fires off the real card, which the engine
//     can only read because the state was cleared BEFORE the event was
//     emitted.
//
// The bounce is a blink enabler in the printed card — return your own
// ETB creature, replay it — and the counter is the compensation for
// spending {3} plus {1}{W} on a 1/1.
//
// "ANOTHER target permanent you control": the Familiar cannot bounce
// itself, which would take the trigger's own source off the
// battlefield before the counter it is about to place.
func init() {
	Register(Spec{
		OracleID:     "7ea0012b-b90b-43d6-bb5b-e4e92452bea7",
		Name:         "Forum Familiar",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Disguise("{1}{W}"),
		},
		Triggered: []game.TriggeredAbility{
			forumFamiliarTurnedFaceUp(),
		},
	})
}

// forumFamiliarTurnedFaceUp is the CR 708.8 trigger, with its target
// clause built per source so "another" can name the Familiar itself.
func forumFamiliarTurnedFaceUp() game.TriggeredAbility {
	t := WhenThisIsTurnedFaceUp("Forum Familiar — bounce a permanent you control", nil)
	t.Targets = TargetPermanent("another target permanent you control", YouControl())
	t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		self := source.InstanceID
		return game.NewTriggeredItem(source, "Forum Familiar — bounce a permanent you control",
			func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) > 0 && item.Targets[0].ID != self {
					if err := g.BounceToHandForEffect(item.Targets[0].ID); err != nil {
						return err
					}
				}
				// CR 608.2: the rest of the ability happens whether or
				// not the target is still legal, and the counter is not
				// conditional on the bounce.
				return g.AddCounterForEffect(self, "+1/+1", 1)
			})
	}
	return t
}
