package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crossroads Watcher — {2}{G} Creature — Kithkin Ranger 3/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Trample
//	 Whenever another creature you control enters, this creature gets
//	 +1/+0 until end of turn."
//
// Mutant Town Musicians without the ability word: one +1/+0 for each
// other creature that enters under your control, tokens included, each
// lasting until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab5f8b8c-e3f7-4265-bbc6-814bbfd4693a",
		Name:            "Crossroads Watcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnotherCreatureEntersUnderYourControl("Crossroads Watcher — +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Crossroads Watcher — +1/+0"}.Apply(ctx)
				}),
		},
	})
}
