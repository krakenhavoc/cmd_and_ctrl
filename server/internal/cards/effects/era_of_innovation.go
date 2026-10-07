package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "355603d1-f62d-407a-8fa9-74e7dca8ecf9",
		Name:         "Era of Innovation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && (c.IsArtifact() || HasSubtype("Artificer")(g, source.Controller, c))
			}, "Era of Innovation — you may pay {1} to get {E}{E}", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Pay {1} to get {E}{E}?",
					OnPay: func(ctx *Context) error {
						return GetEnergy{N: 2}.Apply(ctx)
					},
				}.Apply(ctx)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay six {E}, Sacrifice this enchantment: Draw three cards.",
			Cost:    Plus(PayEnergy(6), SacrificeThis()),
			Purpose: game.Purpose{Draws: 3},
			Effect:  Do(DrawCards{N: 3}),
		}},
	})
}
