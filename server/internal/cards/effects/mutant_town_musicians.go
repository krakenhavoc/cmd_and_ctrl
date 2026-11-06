package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mutant Town Musicians — {2}{R} Creature — Mutant Bard Performer 2/4
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Trample
//	 Alliance — Whenever another creature you control enters, this
//	 creature gets +1/+0 until end of turn."
//
// Alliance is an ability word (CR 207.2c), Slash, Reptile Rampager's
// trigger: one +1/+0 per other creature that enters under your control,
// tokens included, each lasting until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9f318ddc-3dc5-4e43-92b3-3a776e41c56e",
		Name:            "Mutant Town Musicians",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Mutant Town Musicians — alliance: +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Mutant Town Musicians — +1/+0"}.Apply(ctx)
				}),
		},
	})
}
