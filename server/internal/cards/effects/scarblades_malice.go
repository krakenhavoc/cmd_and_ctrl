package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scarblade's Malice — Instant {B} (Lorwyn Eclipsed, #4458):
//
//	"Target creature you control gains deathtouch and lifelink until end of
//	 turn. When that creature dies this turn, create a 2/2 black and green
//	 Elf creature token."
//
// Felonious Rage's shape: the keywords until end of turn, then a delayed
// triggered ability the spell creates as it resolves (CR 603.7a), bound to
// that one object, so a creature that blinks and comes back is a new object
// whose death makes nothing (CR 400.7). The Elf is the spell controller's,
// who controls the delayed trigger (CR 603.7d). A target gone by resolution
// does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66d57851-2844-4b9d-be78-e90fc620b750",
		Name:         "Scarblade's Malice",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch", "lifelink")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !onBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"deathtouch", "lifelink"},
				Label: "Scarblade's Malice — deathtouch and lifelink"}).Apply(ctx); err != nil {
				return err
			}
			return whenThatCreatureDiesThisTurn(ctx, t.ID, "Scarblade's Malice — create a 2/2 black and green Elf", scarbladesMaliceElfBody)
		},
	})
}

// scarbladesMaliceElfBody is the delayed trigger's body: a 2/2 black and
// green Elf for its controller. An on-disk key.
var scarbladesMaliceElfBody = game.SimpleDelayedBody("scarblades-malice/elf", func(g *game.Game, item *game.StackItem) error {
	return CreateToken{Template: TokenCard("2/2 black and green Elf"), N: 1}.Apply(NewContext(g, item))
})
