package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verdurous Gearhulk — Artifact Creature — Construct {3}{G}{G}, 4/4
// (#1658, unblocked by #1656):
//
//	"Trample
//	 When this creature enters, distribute four +1/+1 counters among
//	 any number of target creatures you control."
//
// The counters twin of Fury's fixed-4 "any number" division
// (PutDividedCounters rather than DealDividedDamage), restricted to
// creatures the controller controls — the Gearhulk itself is a legal
// target of its own trigger, printed text and all.
func init() {
	Register(Spec{
		OracleID:        "3765e8bb-e70d-4503-b8dc-1e684e434c18",
		Name:            "Verdurous Gearhulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetCreature("any number of target creatures you control", YouControl()).
				WithCount(0, 0).Dividing(Divide(4)),
			Key: "Verdurous Gearhulk — distribute four +1/+1 counters",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutDividedCounters(NewContext(g, item), game.CounterPlusOne)
			},
		}},
	})
}
