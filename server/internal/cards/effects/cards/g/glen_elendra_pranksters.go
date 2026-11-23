package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glen Elendra Pranksters — Creature — Faerie Wizard {3}{U}, 1/3 (Jumpstart 2022, #5948):
//
//	"Flying
//	 Whenever you cast a spell during an opponent's turn, you may return
//	 target creature you control to its owner's hand."
//
// The trigger fires for a spell its controller casts while another player
// is the active player; the target is chosen as the trigger goes on the
// stack and the "may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4b02df52-d6c0-4735-b2b8-d637cb41f0ba",
		Name:            "Glen Elendra Pranksters",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return ev.Kind == game.EventCast && ev.Actor == source.Controller && !effects.IsActivePlayer(g, source.Controller)
					}, "Glen Elendra Pranksters — you may return target creature you control to its owner's hand",
						effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
							return effects.BounceToHand{Target: id}.Apply(ctx)
						}))),
					"Glen Elendra Pranksters — return that creature to its owner's hand?"),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
