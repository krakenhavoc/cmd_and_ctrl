package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heartlash Cinder — {1}{R} Creature — Elemental Warrior, 1/1:
//
//	"Haste
//	 Chroma — When this creature enters, it gets +X/+0 until end of turn, where
//	 X is the number of red mana symbols in the mana costs of permanents you
//	 control."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "04560bc2-df00-4f75-95a0-a63e2ce760b0",
		Name:            "Heartlash Cinder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Heartlash Cinder — +X/+0 for each red mana symbol", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.S82OldB521ChromaOfPermanents(g, item.Controller, ctx.TextColor("R"))
				if x <= 0 || !effects.OnBattlefield(g, ctx.Source()) {
					return nil
				}
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: x, Label: "Heartlash Cinder — +X/+0"}.Apply(ctx)
			}),
		},
	})
}
