package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moon Girl and Devil Dinosaur — {1}{G}{U} Legendary Creature — Human
// Dinosaur Hero 2/2 (Marvel Super Heroes, #4455):
//
//	"Whenever you draw your second card each turn, until end of turn, Moon
//	 Girl and Devil Dinosaur's base power and toughness become 6/6 and they
//	 gain trample.
//	 Whenever an artifact you control enters, draw a card. This ability
//	 triggers only once each turn."
//
// The first trigger sets the base size at layer 7b and adds trample, both
// until end of turn, to the object that is still the source (a creature that
// left, or came back as a new object, is not touched, CR 400.7); counters and
// pumps still apply on top of the 6/6. The second is a once-per-turn draw on
// any artifact entering under your control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfe0818b-ae13-4174-85c8-24b895d28df0",
		Name:         "Moon Girl and Devil Dinosaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Moon Girl and Devil Dinosaur — base power and toughness 6/6 and trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, ctx.Source()) {
						return nil
					}
					return effects.UntilEndOfTurn(ctx, ctx.Source(), nil,
						"Moon Girl and Devil Dinosaur — base 6/6 and trample until end of turn",
						game.SetBasePowerMod(6), game.SetBaseToughnessMod(6), game.AddKeywordsMod("trample"))
				}),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsArtifact() && !effects.B11TriggeredThisTurn(g, source.InstanceID, moonGirlDrawLabel)
			}, moonGirlDrawLabel, effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

const moonGirlDrawLabel = "Moon Girl and Devil Dinosaur — draw a card"
