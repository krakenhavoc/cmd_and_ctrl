package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Break Down — Instant {2}{G} (Fallout, #5469):
//
//	"Destroy target artifact or enchantment. Create a Junk token. (It's
//	 an artifact with "{T}, Sacrifice this token: Exile the top card of
//	 your library. You may play that card this turn. Activate only as a
//	 sorcery.")"
//
// Nature's Claim's target, then the printed Junk token (JunkToken,
// CR 111.10t). The Junk is made whether or not the target was
// destroyed (an indestructible one survives); with its only target gone
// the spell does not resolve, so there is no Junk (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73094bde-5466-41a4-88f9-a98e67f5ed2e",
		Name:         "Break Down",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard && ctx.IsTargetLegal(item.Targets[0]) {
				if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return CreateToken{Template: JunkToken(), N: 1}.Apply(ctx)
		},
	})
}
