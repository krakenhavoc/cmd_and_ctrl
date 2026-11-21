package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xavier Sal, Infested Captain — Legendary Creature — Human Fungus Pirate
// {B}{G}{U}, 3/3 (#5368):
//
//	"{T}, Remove a counter from another permanent you control: Populate.
//	 Activate only as a sorcery.
//	 {T}, Sacrifice another creature: Proliferate. Activate only as a
//	 sorcery."
//
// Both costs are chosen as they are paid, not targeted (CR 602.2b). The
// first removes a counter of any kind from a permanent its controller
// controls other than Xavier Sal (the counter cost's any-kind form, Burdened
// Stoneback's, from a chosen permanent as Heart of Kiran's), then populates:
// with no creature token to copy it does nothing. The second sacrifices
// another creature and proliferates (Contagion Clasp's primitive). Both tap,
// so they wait out summoning sickness (CR 302.6), and both are sorcery speed.
//
// No simplification.
func init() {
	removeCounter := effects.RemoveCountersFrom("", 1, "another permanent you control")
	removeCounter.RemoveCounters.From = effects.Another(removeCounter.RemoveCounters.From)
	effects.Register(effects.Spec{
		OracleID:     "50a9c647-47ba-4732-852a-f0ab2b4549c9",
		Name:         "Xavier Sal, Infested Captain",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:        "{T}, Remove a counter from another permanent you control: Populate. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.TapCost(), removeCounter),
				SorcerySpeed: true,
				Effect:       effects.Do(effects.Populate{}),
			},
			{
				Label:        "{T}, Sacrifice another creature: Proliferate. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.TapCost(), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Proliferate{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
