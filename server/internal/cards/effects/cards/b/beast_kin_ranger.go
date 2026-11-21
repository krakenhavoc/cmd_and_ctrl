package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beast-Kin Ranger — Creature — Elf Ranger {2}{G}, 3/3 (Foundations,
// #4465):
//
//	"Trample
//	 Whenever another creature you control enters, this creature gets
//	 +1/+0 until end of turn."
//
// Crossroads Watcher's text: one +1/+0 for each other creature that
// enters under your control, tokens included, each lasting until end of
// turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "db794af5-ba2e-4b10-b213-e481e8a50432",
		Name:            "Beast-Kin Ranger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnotherCreatureEntersUnderYourControl("Beast-Kin Ranger — +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Beast-Kin Ranger — +1/+0"}.Apply(ctx)
				}),
		},
	})
}
