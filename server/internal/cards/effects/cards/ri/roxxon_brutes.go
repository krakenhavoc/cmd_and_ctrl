package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roxxon Brutes — Creature — Human Berserker Villain {4}{B}, 4/4 (Marvel
// Super Heroes, #4455):
//
//	"Menace
//	 Whenever you draw your second card each turn, put a +1/+1 counter on
//	 target creature.
//	 Basic landcycling {2}"
//
// Lakeshore Apothecary's second-draw watch with the counter on a target
// creature (any creature, chosen when the trigger goes on the stack), and
// Apex Witchstalker's basic landcycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b92c2da9-eaa4-404b-8114-5527eec6390e",
		Name:            "Roxxon Brutes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
				}, "Roxxon Brutes — put a +1/+1 counter on target creature", effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
