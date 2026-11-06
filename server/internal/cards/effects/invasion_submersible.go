package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion Submersible — Artifact — Vehicle {2}{U}, 0/0:
//
//	"When this Vehicle enters, return up to one other target nonland
//	 permanent to its owner's hand.
//	 Exhaust — Waterbend {3}: This Vehicle becomes an artifact creature.
//	 Put three +1/+1 counters on it. (While paying a waterbend cost, you
//	 can tap your artifacts and creatures to help. Each one pays for {1}.
//	 Activate each exhaust ability only once.)"
//
// The entry is Bespoke Bō's "up to one other" bounce. The exhaust is
// Rangers' Refueler's animation with Avatar Kuruk's waterbend cost: the
// {3} is the mana component, and each artifact or creature tapped toward
// it pays {1} (CR 701.67a). The animation has no duration (CR 611.2a) and
// is pinned to this object, so a Vehicle that left and came back is a new
// object, un-animated, whose exhaust can be activated again (CR 400.7, 702.177a,
// the ruling). It prints no crew.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "97369152-3b85-4cf1-aeb9-b3a8f1c37611",
		Name:         "Invasion Submersible",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Invasion Submersible — return up to one other target nonland permanent to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						return bounceTheTarget(item, NewContext(g, item))
					}),
				Another(TargetPermanent("up to one other target nonland permanent", Nonland()).WithCount(0, 1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — Waterbend {3}: This Vehicle becomes an artifact creature. Put three +1/+1 counters on it.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
			Exhaust: true,
			Cost:    WaterbendCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if sourceIsNewObject(g, item) {
					return nil
				}
				if err := (BecomeArtifactCreature{
					Label: "Invasion Submersible — becomes an artifact creature",
				}).Apply(ctx); err != nil {
					return err
				}
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
			},
		}},
	})
}
