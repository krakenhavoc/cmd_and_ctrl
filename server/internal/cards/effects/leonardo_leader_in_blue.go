package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonardo, Leader in Blue — Legendary Creature — Mutant Ninja Turtle
// {W}, 2/1 (TMT, #4457):
//
//	"Sneak {3}{W}{W} (… He enters tapped and attacking.)
//	 When Leonardo enters, if his sneak cost was paid, creatures you
//	 control get +2/+0 until end of turn.
//	 {1}{W}: Leonardo gains first strike until end of turn."
//
// Sneak is Sneak(cost) (CR 702.190). The intervening if (CR 603.4)
// reads the entering permanent's cast record, as Diminisher Witch reads
// bargain; the record cannot change while he is on the battlefield, so
// the resolution check agrees. The pump fixes its set of creatures as
// it resolves (CR 611.2c). The activated ability is Fangblade Brigand's
// firebreathing shape without the +1/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "348c7e22-2179-41b8-a5d2-8dd488e7e66d",
		Name:             "Leonardo, Leader in Blue",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{3}{W}{W}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, tmtB11EnteredSneaked),
				"Leonardo, Leader in Blue — creatures you control get +2/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl()), Power: 2,
						Label: "Leonardo, Leader in Blue — creatures you control get +2/+0",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Leonardo gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{W}"),
			Effect:  thisCreatureUntilEOT("Leonardo, Leader in Blue — first strike", 0, 0, "first strike"),
		}},
	})
}
