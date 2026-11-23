package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pack Attack — Instant {2}{R} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"Attacking creatures get +X/+0 until end of turn, where X is the number of
//	 players being attacked.
//	 Draw a card."
//
// X is the number of distinct defending players among the attacking creatures
// (a creature attacking a planeswalker attacks that planeswalker's controller),
// counted once as the spell resolves; the pump is fixed for the creatures that
// are attacking then (CR 611.2c). With no attackers X is 0 and the spell is a
// cantrip.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84b385cd-9dfd-4513-af3c-20fc1cc8fea0",
		Name:         "Pack Attack",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			seen := map[uuid.UUID]bool{}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.AttackingTarget != uuid.Nil {
					if p := ctx.Game.DefendingPlayerForAttackForEffect(c.AttackingTarget); p != uuid.Nil {
						seen[p] = true
					}
				}
			}
			if x := len(seen); x > 0 {
				if err := (effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: x, Label: "Pack Attack — attacking creatures get +X/+0"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
