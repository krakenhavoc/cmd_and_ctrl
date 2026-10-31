package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "b92c2da9-eaa4-404b-8114-5527eec6390e",
		Name:            "Roxxon Brutes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b41DrewYourSecondCardThisTurn(ev, source, g)
				}, "Roxxon Brutes — put a +1/+1 counter on target creature", PlusOneCounterOnTheTarget),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
