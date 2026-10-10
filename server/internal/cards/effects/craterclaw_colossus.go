package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Craterclaw Colossus — Artifact Creature — Beast Construct {4}{R}{R}{R}, 5/5:
//
//	"Haste
//	 When this creature enters, creatures you control gain trample and
//	 get +X/+0 until end of turn, where X is the number of artifacts
//	 you control."
//
// The set of creatures is locked as the trigger resolves (CR 611.2c),
// and X is read at the same moment, so a creature that arrives later
// in the turn gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "71355b51-e70e-47b0-8460-933af3567141",
		Name:            "Craterclaw Colossus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Craterclaw Colossus — creatures you control gain trample and get +X/+0",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					mine := And(Creature(), YouControl())
					if err := (GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"},
						Label: "Craterclaw Colossus — trample"}).Apply(ctx); err != nil {
						return err
					}
					return BoostUntilEOT{Match: mine, Power: rfCreatureAArtifactsYouControl(g, item.Controller),
						Label: "Craterclaw Colossus — +X/+0"}.Apply(ctx)
				}),
		},
	})
}
