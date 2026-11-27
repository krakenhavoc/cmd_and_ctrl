package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Searing Rays — Sorcery {2}{R} (Invasion, S82 #5616):
//
//	"Choose a color. Searing Rays deals damage to each player equal to the number
//	 of creatures of that color that player controls."
//
// The colour is named as the spell resolves; then each player is counted and
// dealt damage as one instance, a player with none taking nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd6730e2-3904-43cf-90c6-e3d9fe210bc5",
		Name:         "Searing Rays",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseColorTail(game.ColorForBenefit, ctx.Game, item.Controller, item.SourceCardID,
				"Searing Rays — choose a color", ctx.ItemTail(searingRaysChosen))
			return nil
		},
	})
}

// searingRaysChosen deals each player damage equal to the number of creatures
// of the chosen colour they control. An on-disk key.
var searingRaysChosen = effects.CardTail("then/searing-rays-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, p := range ctx.Game.Seats {
			if p == nil || p.Eliminated {
				continue
			}
			n := 0
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.OfColor(color))) {
				if c.Controller == p.ID {
					n++
				}
			}
			if n <= 0 {
				continue
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: p.ID, Amount: n}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
})
