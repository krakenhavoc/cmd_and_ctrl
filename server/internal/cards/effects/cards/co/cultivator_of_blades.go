package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cultivator of Blades — Creature — Elf Artificer {3}{G}{G}, 1/1 (Kaldheim
// Commander, #5506):
//
//	"Fabricate 2
//	 Whenever this creature attacks, you may have other attacking
//	 creatures get +X/+X until end of turn, where X is this creature's
//	 power."
//
// Fabricate is the shared keyword constructor. X is read as the trigger
// resolves (CR 608.2h, last-known power if the Cultivator has left), and
// the set of attackers is locked then (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca4c0945-7035-47dc-aba8-fbda351d9b7b",
		Name:         "Cultivator of Blades",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Fabricate("Cultivator of Blades", 2),
			effects.WheneverThisAttacks("Cultivator of Blades — you may have other attacking creatures get +X/+X, X is its power",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Player:   item.Controller,
						Question: "Cultivator of Blades — have other attacking creatures get +X/+X?",
						YesLabel: "Pump them",
						NoLabel:  "No",
						Then:     cultivatorOfBladesPump,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// cultivatorOfBladesPump is the "yes". An on-disk key: never renamed,
// never reused.
var cultivatorOfBladesPump = effects.MayYes("confirm/cultivator-of-blades-pump", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	x := effects.SourcePowerNow(ctx)
	return effects.BoostUntilEOT{
		Match:     effects.And(effects.AttackingCreature(), effects.OtherThan(ctx.Source())),
		Power:     x,
		Toughness: x,
		Label:     "Cultivator of Blades",
	}.Apply(ctx)
})
