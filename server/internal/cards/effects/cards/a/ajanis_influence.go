package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani's Influence — Sorcery {2}{W}{W} (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"Put two +1/+1 counters on target creature.
//	 Look at the top five cards of your library. You may reveal a white card
//	 from among them and put it into your hand. Put the rest on the bottom of
//	 your library in a random order."
//
// Two counters on the target, then Skalla Wolf's look-at-five for a white
// card. With its one target illegal the whole spell does nothing (CR
// 608.2b), the look included.
//
// No simplification.
func init() {
	look := effects.LookAtTopThenMayTakeToHand(5, effects.OfColor("W"), 1,
		"Ajani's Influence — you may reveal a white card and put it into your hand")
	effects.Register(effects.Spec{
		OracleID:     "2a3e3f3b-1f74-4ac7-b5ac-7c3cb9cc38a6",
		Name:         "Ajani's Influence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return look(ctx.Game, item)
		},
	})
}
