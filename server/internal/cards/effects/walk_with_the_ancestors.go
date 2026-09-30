package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Walk with the Ancestors — Sorcery {4}{G}:
//
//	"Return up to one target permanent card from your graveyard to your
//	 hand. Discover 4."
//
// The return goes first, as printed, so a card it returns is in hand
// before the discover walks the library. The discover happens with or
// without a target. Discover is ADR 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "ce657a2e-461b-45fd-ab98-35b5c8512777",
		Name:         "Walk with the Ancestors",
		Completeness: CompletenessFull,
		Discovers:    true,
		Targets:      TargetCardInGraveyard("up to one target permanent card from your graveyard", YouOwn(), Permanent()).WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return Discover{N: 4}.Apply(ctx)
		},
	})
}
