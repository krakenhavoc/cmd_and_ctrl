package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Go for Blood — Sorcery {1}{R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature you control fights target creature you don't control.
//	 (Each deals damage equal to its power to the other.)
//	 Cycling {1} ({1}, Discard this card: Draw a card.)"
//
// Two clauses, judged each on its own (CR 601.2c, 608.2b), and the
// shared fight body (slotZeroFightsSlotOne): with either gone, no
// fight. Cycling from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11cf103c-f8a3-4892-942d-2c4472dd5716",
		Name:         "Go for Blood",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SlotZeroFightsSlotOne(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{1}")},
	})
}
