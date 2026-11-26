package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invert the Skies — Instant {3}{G/U}:
//
//	"Creatures your opponents control lose flying until end of turn if {G} was
//	 spent to cast this spell, and creatures you control gain flying until end
//	 of turn if {U} was spent to cast this spell. (Do both if {G}{U} was
//	 spent.)"
//
// River's Grasp's mana-spent reading for each half (ReadsOwnPayment; a copy
// has no mana spent and does neither). The creatures each half names are
// those on the battlefield as the spell resolves (CR 611.2c), and the first
// half is applied first, so a creature that is both an opponent's and
// yours-by-theft is fixed by the printed order.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ea5ef11d-635f-4f62-908d-6d14465eb921",
		Name:           "Invert the Skies",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"G", "U"},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.ManaSpentOfColor("G") > 0 {
				if err := effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.OpponentControls()),
					"Invert the Skies — creatures lose flying until end of turn", game.RemoveKeywordsMod("flying")); err != nil {
					return err
				}
			}
			if ctx.ManaSpentOfColor("U") > 0 {
				return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"flying"},
					Label: "Invert the Skies — creatures you control gain flying until end of turn"}.Apply(ctx)
			}
			return nil
		},
	})
}
