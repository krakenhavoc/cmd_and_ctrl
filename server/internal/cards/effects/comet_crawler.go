package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Comet Crawler — Creature — Insect Horror {2}{B}, 2/3 (Edge of Eternities,
// #4461):
//
//	"Lifelink
//	 Whenever this creature attacks, you may sacrifice another creature or
//	 artifact. If you do, this creature gets +2/+0 until end of turn."
//
// Beetle-Headed Merchants' attack trigger (maySacrificeOne, asked at
// resolution over what you control then, the source excluded) with a pump
// in place of the draw and the counter. The pump goes on this creature only
// while it is still the object that attacked (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0c1b6305-8863-43e0-b50d-ae1c5e31be55",
		Name:            "Comet Crawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Comet Crawler — you may sacrifice another creature or artifact to get +2/+0",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Comet Crawler — sacrifice another creature or artifact to get +2/+0 until end of turn?",
						Or(Creature(), Artifact()), true, cometCrawlerSacrificed)
				}),
		},
	})
}

var cometCrawlerSacrificed = RegisterSacrificedOne("prompt-run/comet-crawler-pump", func(ctx *Context, _ uuid.UUID) error {
	if sourceIsNewObject(ctx.Game, ctx.Item) || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return BoostUntilEOT{Target: ctx.Source(), Power: 2, Label: "Comet Crawler — +2/+0"}.Apply(ctx)
})
