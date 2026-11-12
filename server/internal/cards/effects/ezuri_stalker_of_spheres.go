package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ezuri, Stalker of Spheres — Legendary Creature — Phyrexian Elf Warrior
// {2}{G}{U}, 3/3 (Phyrexia: All Will Be One, #5653):
//
//	"When Ezuri enters, you may pay {3}. If you do, proliferate twice.
//	 Whenever you proliferate, draw a card."
//
// The enters trigger asks for {3} as it resolves (MayPay); paid, the two
// proliferates are each their own choice, the second asked once the
// first's counters have landed (proliferateTimesKeyed). Each one is a
// proliferate of its own (CR 701.34a), so the second ability triggers
// twice, and once more for any other proliferate its controller makes,
// one that chose nothing included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cffec16f-373b-43e7-a04c-4452987d7e02",
		Name:         "Ezuri, Stalker of Spheres",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ezuri, Stalker of Spheres — you may pay {3}; if you do, proliferate twice",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{3}",
						Question: "Ezuri, Stalker of Spheres — pay {3} to proliferate twice?",
						Then:     ezuriStalkerOfSpheresPaid,
					}.Apply(NewContext(g, item))
				}),
			WheneverYouProliferate("Ezuri, Stalker of Spheres — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// ezuriStalkerOfSpheresPaid is the paid half of Ezuri's enters trigger:
// proliferate twice. An on-disk key: never renamed, never reused.
var ezuriStalkerOfSpheresPaid = PayUnlessBranches("pay-unless/ezuri-stalker-of-spheres-proliferate-twice",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return proliferateTimesKeyed(ctx.Game, ctx.Controller(), ctx.Source(), 2)
	}, nil)
