package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Linvala, the Preserver — Legendary Creature — Angel {4}{W}{W}, 5/5:
//
//	"Flying
//	 When Linvala enters, if an opponent has more life than you, you
//	 gain 5 life.
//	 When Linvala enters, if an opponent controls more creatures than
//	 you, create a 3/3 white Angel creature token with flying."
//
// Two separate intervening-"if" triggers (CR 603.4): each is checked as
// Linvala enters and again as its ability resolves, and "an opponent"
// may be a different opponent each time. Beza's comparisons.
func init() {
	gain := effects.WhenThisEnters("Linvala, the Preserver — gain 5 life", effects.Do(effects.GainLife{Amount: 5}))
	gain.AppliesTo = effects.AllOf(gain.AppliesTo, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.B39AnOpponentHasMoreLife(g, source.Controller)
	})
	angel := effects.WhenThisEnters("Linvala, the Preserver — create a 3/3 white Angel creature token with flying",
		effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 white Angel with flying"), N: 1}))
	angel.AppliesTo = effects.AllOf(angel.AppliesTo, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.B39AnOpponentControlsMoreCreatures(g, source.Controller)
	})
	effects.Register(effects.Spec{
		OracleID:        "64f5689d-504d-40e9-bdb5-36dcefd257e2",
		Name:            "Linvala, the Preserver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "an opponent has more life than you", Test: func(ctx *effects.Context) bool {
				return effects.B39AnOpponentHasMoreLife(ctx.Game, ctx.Controller())
			}}, gain),
			effects.InterveningIf(effects.CondBoard{Name: "an opponent controls more creatures than you", Test: func(ctx *effects.Context) bool {
				return effects.B39AnOpponentControlsMoreCreatures(ctx.Game, ctx.Controller())
			}}, angel),
		},
	})
}
