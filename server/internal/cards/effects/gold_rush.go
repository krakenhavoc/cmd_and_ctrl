package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gold Rush — Instant {1}{G} (Outlaws of Thunder Junction, #4468):
//
//	"Create a Treasure token. Until end of turn, up to one target
//	 creature gets +2/+2 for each Treasure you control."
//
// Two steps in the printed order: the Treasure is made first, so it
// counts. The count is taken once, as the pump is applied (CR 608.2h),
// and the bonus is then fixed for the turn: sacrificing a Treasure
// afterwards does not shrink the creature. With no target chosen, or a
// target gone by resolution, the spell still makes its Treasure.
//
// The steps are a ResolveSteps sequence, so a CR 616 ordering prompt the
// token creation opens (two token replacements) holds the pump until it
// is answered and the new Treasures are counted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45687a24-e8cb-4864-8721-410eb8320aab",
		Name:         "Gold Rush",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to one target creature").WithCount(0, 1),
		Purpose:      game.Purpose{Tokens: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0))},
		OnResolve:    ResolveSteps(CreateToken{Template: TreasureToken(), N: 1}, Instruction(goldRushPump)),
	})
}

// goldRushPump is "up to one target creature gets +2/+2 for each
// Treasure you control" until end of turn, counted now.
func goldRushPump(ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return hasSubtype(c, "Treasure") })
	return BoostUntilEOT{Target: id, Power: 2 * n, Toughness: 2 * n, Label: "Gold Rush — +2/+2 for each Treasure"}.Apply(ctx)
}
