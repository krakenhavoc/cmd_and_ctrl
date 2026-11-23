package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeskai Ascendancy — Enchantment {U}{R}{W} (S82 GAP, #5948):
//
//	"Whenever you cast a noncreature spell, creatures you control get +1/+1
//	 until end of turn. Untap those creatures.
//	 Whenever you cast a noncreature spell, you may draw a card. If you do,
//	 discard a card."
//
// Two triggers, one per printed ability. The creatures you control as the first
// resolves are the ones pumped and untapped (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3c802ba-dfbb-4f63-ae9a-e29c2e30013d",
		Name:         "Jeskai Ascendancy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou,
				"Jeskai Ascendancy — creatures you control get +1/+1 until end of turn; untap those creatures",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if err := (effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
						Label: "Jeskai Ascendancy — +1/+1",
					}).Apply(ctx); err != nil {
						return err
					}
					return effects.UntapAllCreaturesYouControl{}.Apply(ctx)
				}))),
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou,
				"Jeskai Ascendancy — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Jeskai Ascendancy — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
