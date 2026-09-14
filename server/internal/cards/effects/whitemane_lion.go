package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whitemane Lion — Creature — Cat {1}{W}, 2/2 (EDHREC rank 3786):
//
//	"Flash
//	 When this creature enters, return a creature you control to its
//	 owner's hand."
//
// The two-mana ETB re-buy: flash it in, bounce the creature whose
// enters trigger you want again — or bounce the Lion itself to save
// it from a wipe, which is a real and printed line. Flash rides
// PrintedKeywords; the bounce is a mandatory ETB trigger.
//
// Sandbox simplification, declared (the Azorius Chancery / Time Wipe
// posture): "return a creature you control" is a resolution-time
// choice, not a target, and the pick_target prompt is the one picker
// the engine has for choosing among permanents — so the creature is
// chosen as a target when the trigger goes on the stack. Weaker than
// printed on two counts: opponents see the choice before the trigger
// resolves, and a creature you control with shroud cannot be the one
// returned. It can never fizzle — the Lion itself is always a legal
// answer.
func init() {
	Register(Spec{
		OracleID:     "e8d6084b-9b72-438e-a30a-851b888f3e4d",
		Name:         "Whitemane Lion",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick the creature to return when the trigger goes on the stack rather than as it resolves, so opponents can respond to the choice, and a creature of yours with shroud can't be picked.",
		},
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetCreature("a creature you control to return to its owner's hand", YouControl()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Whitemane Lion — return a creature you control to its owner's hand",
					b36BounceChosenCreature)
			},
		}},
	})
}
