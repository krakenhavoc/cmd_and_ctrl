package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pest Infestation — Sorcery {X}{X}{G} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Destroy up to X target artifacts and/or enchantments. Create twice X
//	 1/1 black and green Pest creature tokens with "When this token dies, you
//	 gain 1 life.""
//
// "Up to X targets" is CountFromX with UpToX (March of Swirling Mist): the
// announced X is the ceiling and fewer targets are legal. The cost is
// {X}{X}{G}, so one announced X pays twice (CR 107.3). Each target still legal
// as the spell resolves is destroyed, then the tokens are made whether or not
// anything was (they are not conditional on the destruction). The Pest is the
// catalog's, with its own dies trigger. X is 0 on the stack of a spell that
// was cast for {G}: nothing is destroyed and no token made, so the spell
// declares XMatters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b704798-0c69-4c18-ac7e-42933ce90028",
		Name:         "Pest Infestation",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      socB10PestInfestationTargets(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, id := range legalTargetCards(item, ctx.Game) {
				if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return CreateToken{Template: PestToken(), N: 2 * ctx.X()}.Apply(ctx)
		},
	})
}

// socB10PestInfestationTargets is "up to X target artifacts and/or
// enchantments".
func socB10PestInfestationTargets() *game.TargetSpec {
	spec := TargetPermanent("up to X target artifacts and/or enchantments", Or(Artifact(), Enchantment()))
	spec.CountFromX = true
	spec.UpToX = true
	return spec
}
