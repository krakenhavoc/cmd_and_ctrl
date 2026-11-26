package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Increasing Confusion — Sorcery {X}{U}:
//
//	"Target player mills X cards. If this spell was cast from a graveyard,
//	 that player mills twice that many cards instead.
//	 Flashback {X}{U}"
//
// The doubling reads CastFromZone, as the card says ("cast from a
// graveyard"), not the cost paid. X is the one chosen for the cast, in
// either zone.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "11293f96-3263-45e5-89fc-850df6b25a63",
		Name:             "Increasing Confusion",
		Completeness:     effects.CompletenessFull,
		XMatters:         true,
		Targets:          effects.TargetPlayer("target player"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{X}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			p := effects.TargetedPlayerOf(ctx)
			if p == uuid.Nil {
				return nil
			}
			n := ctx.X()
			if item.CastFromZone == game.ZoneGraveyard {
				n *= 2
			}
			return effects.MillCards{Player: p, N: n}.Apply(ctx)
		},
	})
}
