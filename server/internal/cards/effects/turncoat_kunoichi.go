package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turncoat Kunoichi — Creature — Mutant Ninja Fox {2}{W}, 3/2 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"Sneak {2}{W}{B} (You may cast this spell for {2}{W}{B} if you also
//	 return an unblocked attacker you control to hand during the declare
//	 blockers step. It enters tapped and attacking.)
//	 When this creature enters, choose target creature an opponent
//	 controls. Exile that creature until this creature leaves the
//	 battlefield. If this creature's sneak cost was paid, instead exile the
//	 chosen creature."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The exile is
// Brutal Cathar's "until this creature leaves" (exileChosenTargetUntilThisLeaves),
// and "if this creature's sneak cost was paid" is read off the entering
// object's cast record (ObjectSnapshot.AltCost), as Adipose Offspring reads
// emerge, so it is the same answer if the Kunoichi has left by the time the
// trigger resolves. Sneaked in, the creature is exiled for good; cast for
// its mana cost or put onto the battlefield, it comes back when the
// Kunoichi leaves.
//
// No simplification.
func init() {
	untilLeaves := exileChosenTargetUntilThisLeaves("Turncoat Kunoichi — the exiled creature returns when Turncoat Kunoichi leaves the battlefield")
	Register(Spec{
		OracleID:         "0f35cf85-6783-4d69-b5e6-a81f7c1f21e7",
		Name:             "Turncoat Kunoichi",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{2}{W}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Turncoat Kunoichi — exile target creature an opponent controls until this creature leaves the battlefield (for good if its sneak cost was paid)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if obj := ctx.Trigger().Object; obj != nil && sneakCostWasPaid(obj.AltCost) {
						return ExileChosenTarget{}.Apply(ctx)
					}
					return untilLeaves(g, item)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
