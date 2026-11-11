package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	removeCounter := RemoveCountersFrom("", 1, "another permanent you control")
	removeCounter.RemoveCounters.From = Another(removeCounter.RemoveCounters.From)
	Register(Spec{
		OracleID:     "50a9c647-47ba-4732-852a-f0ab2b4549c9",
		Name:         "Xavier Sal, Infested Captain",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:        "{T}, Remove a counter from another permanent you control: Populate. Activate only as a sorcery.",
				Cost:         Plus(TapCost(), removeCounter),
				SorcerySpeed: true,
				Effect:       Do(Populate{}),
			},
			{
				Label:        "{T}, Sacrifice another creature: Proliferate. Activate only as a sorcery.",
				Cost:         Plus(TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Proliferate{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
