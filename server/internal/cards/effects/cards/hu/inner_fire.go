package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inner Fire — Sorcery {3}{R} (Saviors of Kamigawa, #5601):
//
//	"Add {R} for each card in your hand."
//
// Jeska's Will's count form of AddMana. The hand is counted as the spell
// resolves (CR 608.2h), with Inner Fire already off it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27348fe5-21d4-40f7-9feb-bba2c2321008",
		Name:         "Inner Fire",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.SokB1CardsInYourHand(ctx)
			if n <= 0 {
				return nil
			}
			return effects.AddMana{Player: ctx.Controller(), Amount: effects.ManaN(n, "R")}.Apply(ctx)
		},
	})
}
