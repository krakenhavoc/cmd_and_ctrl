package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bespoke Bō — {2}{U} Artifact — Equipment (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"When this Equipment enters, return up to one other target nonland
//	 permanent to its owner's hand.
//	 Equipped creature gets +2/+1 and has vigilance.
//	 Equip {3}"
//
// The enters trigger is Disruptor of Currents's: "other" is object identity,
// "up to one" lets it target nothing, and the target is checked again as the
// trigger resolves (CR 608.2b). The statics are Cathar's Shield's pump and
// keyword grant on the equipped creature, and the equip is CR 702.6a's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "679480aa-f9ad-4a6c-bd79-db9f9f75433d",
		Name:         "Bespoke Bō",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Bespoke Bō — return up to one other target nonland permanent to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						return bounceTheTarget(item, NewContext(g, item))
					}),
				Another(TargetPermanent("up to one other target nonland permanent", Nonland()).WithCount(0, 1))),
		},
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
			GrantToAttached("vigilance"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
