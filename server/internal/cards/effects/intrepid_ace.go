package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intrepid Ace — Creature — Human Pilot {R}, 2/1 (Marvel Super Heroes
// Commander, #4455): "This creature gets +2/+0 as long as it isn't
// attacking or blocking."
//
// One self-only layer 7c static. It reads attacking status (so it
// declares DependsOnAttackingStatus, the same as Ohran Frostfang's
// family) and the blocking pointer, which the layer listener also
// re-reads on a block declaration (EventBlock, layer_listener.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb8889c8-8ee3-4e27-9f47-ad1021ff61bb",
		Name:         "Intrepid Ace",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:                    "Intrepid Ace gets +2/+0 as long as it isn't attacking or blocking.",
				Layer:                    game.Layer7PT,
				SubLayer:                 game.SubLayer7C_Modify,
				DependsOnAttackingStatus: true,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return selfOnly(target, g, source) &&
						target.AttackingTarget == uuid.Nil && target.BlockingTarget == uuid.Nil
				},
				Apply: pumpBy(2, 0),
			},
		},
	})
}
