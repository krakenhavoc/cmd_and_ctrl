package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "7c3982c5-6af3-4937-8d1f-e006f4e89342",
		Name:         "Kill-Zone Acrobat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kill-Zone Acrobat — you may sacrifice another creature or artifact to gain flying",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Kill-Zone Acrobat — sacrifice another creature or artifact to gain flying until end of turn?",
						Or(Creature(), Artifact()), true, killZoneAcrobatFlying)
				}),
		},
	})
}

// killZoneAcrobatFlying is "if you do, this creature gains flying until end
// of turn". An on-disk key.
var killZoneAcrobatFlying = RegisterSacrificedOne("prompt-run/kill-zone-acrobat-flying", func(ctx *Context, _ uuid.UUID) error {
	if sourceIsNewObject(ctx.Game, ctx.Item) || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"}, Label: "Kill-Zone Acrobat — flying"}.Apply(ctx)
})
