package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shuri, the Black Panther — Legendary Creature — Human Noble Hero {G}{W},
// 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Lifelink
//	 Whenever Shuri attacks, draw a card if you control three or more
//	 artifacts. Then if you control six or more artifacts, creatures you
//	 control get +2/+2 until end of turn."
//
// Both counts are taken as the ability resolves, in printed order: the
// draw (which can put an artifact into play only through another
// trigger, never itself) and then the six-artifact check, so each
// "if" reads the board at the point its sentence is reached (CR 608.2c).
// The pump is Black Panther, Vanguard's mass boost, evaluated once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1ea7b298-be87-4032-9f45-37f9b39ce3fa",
		Name:            "Shuri, the Black Panther",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Shuri — draw a card if you control three or more artifacts; then +2/+2 with six", shuriAttacks),
		},
	})
}

func shuriAttacks(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	artifacts := func() int {
		return effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.IsArtifact() })
	}
	if artifacts() >= 3 {
		if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if artifacts() < 6 {
		return nil
	}
	return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 2,
		Label: "Shuri — creatures you control get +2/+2"}.Apply(ctx)
}
