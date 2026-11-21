package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kill-Zone Acrobat — Creature — Human Soldier {2}{B}, 3/2 (The Brothers'
// War, S82):
//
//	"Whenever this creature attacks, you may sacrifice another creature or
//	 artifact. If you do, this creature gains flying until end of turn."
//
// Automated Warfare System's attack trigger (maySacrificeOne, asked as it
// resolves, the source excluded). The flying goes to the Acrobat only while
// it is still the object that attacked (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c3982c5-6af3-4937-8d1f-e006f4e89342",
		Name:         "Kill-Zone Acrobat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Kill-Zone Acrobat — you may sacrifice another creature or artifact to gain flying",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Kill-Zone Acrobat — sacrifice another creature or artifact to gain flying until end of turn?",
						effects.Or(effects.Creature(), effects.Artifact()), true, killZoneAcrobatFlying)
				}),
		},
	})
}

// killZoneAcrobatFlying is "if you do, this creature gains flying until end
// of turn". An on-disk key.
var killZoneAcrobatFlying = effects.RegisterSacrificedOne("prompt-run/kill-zone-acrobat-flying", func(ctx *effects.Context, _ uuid.UUID) error {
	if effects.SourceIsNewObject(ctx.Game, ctx.Item) || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"}, Label: "Kill-Zone Acrobat — flying"}.Apply(ctx)
})
