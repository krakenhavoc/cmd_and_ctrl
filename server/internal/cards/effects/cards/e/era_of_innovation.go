package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Era of Innovation — Enchantment {1}{U}:
//
//	"Whenever an artifact or Artificer you control enters, you may pay
//	 {1}. If you do, you get {E}{E} (two energy counters).
//	 Pay six {E}, Sacrifice this enchantment: Draw three cards."
//
// The {1} is paid as the trigger resolves (CR 118.12), through the
// MayPay prompt; declining gets nothing. ADR 0129 §2 (#1995): "Pay six
// {E}" is the energy cost component, paid with the sacrifice as the
// ability is activated.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "355603d1-f62d-407a-8fa9-74e7dca8ecf9",
		Name:         "Era of Innovation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && (c.IsArtifact() || effects.HasSubtype("Artificer")(g, source.Controller, c))
			}, "Era of Innovation — you may pay {1} to get {E}{E}", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Pay {1} to get {E}{E}?",
					Then:     eraOfInnovationEnergy,
				}.Apply(ctx)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay six {E}, Sacrifice this enchantment: Draw three cards.",
			Cost:    effects.Plus(effects.PayEnergy(6), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 3},
			Effect:  effects.Do(effects.DrawCards{N: 3}),
		}},
	})
}

// eraOfInnovationEnergy is the paid half, keyed (#4264).
var eraOfInnovationEnergy = effects.PayUnlessBranches("pay-unless/era-of-innovation-get-energy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return effects.GetEnergy{N: 2}.Apply(ctx) }, nil)
