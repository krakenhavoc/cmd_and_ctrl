package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renegade Wheelsmith — Creature — Dwarf Pilot {1}{R}{W}, 3/2 (Aether
// Revolt, S82 older sets, #5536):
//
//	"Whenever this creature becomes tapped, target creature can't block
//	 this turn."
//
// It becomes tapped when it attacks without vigilance, when it is tapped
// for a cost (crewing a Vehicle) and when an effect taps it; an attack is
// watched beside the tap event because a tapped attacker is not announced
// as a tap by the declaration. The target is chosen as the trigger goes on
// the stack and re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3caee32-25b7-4ef7-9f9a-e31d08720086",
		Name:         "Renegade Wheelsmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Renegade Wheelsmith — target creature can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Renegade Wheelsmith — can't block this turn"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature")),
		},
	})
}
