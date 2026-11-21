package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walk with the Ancestors — Sorcery {4}{G}:
//
//	"Return up to one target permanent card from your graveyard to your
//	 hand. Discover 4."
//
// The return goes first, as printed, so a card it returns is in hand
// before the discover walks the library. The discover happens with or
// without a target. Discover is ADR 0099's (game/discover.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce657a2e-461b-45fd-ab98-35b5c8512777",
		Name:         "Walk with the Ancestors",
		Purpose:      game.Purpose{Discover: 4},
		Completeness: effects.CompletenessFull,
		Discovers:    true,
		Targets:      effects.TargetCardInGraveyard("up to one target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()).WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Discover{N: 4}.Apply(ctx)
		},
	})
}
