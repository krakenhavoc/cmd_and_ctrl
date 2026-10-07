package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flourishing Defenses — Enchantment {4}{G}:
//
//	"Whenever a -1/-1 counter is put on a creature, you may create a
//	 1/1 green Elf Warrior creature token."
//
// #1841: once per counter, on any creature, whoever controls it
// (WheneverACounterIsPutOnACreature).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1b26eb1-6337-42d6-a475-53ad6427f00f",
		Name:         "Flourishing Defenses",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WheneverACounterIsPutOnACreature(game.CounterMinusOne, "Flourishing Defenses — you may create a 1/1 green Elf Warrior",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: TokenCard("1/1 green Elf Warrior"), N: 1}.Apply(NewContext(g, item))
				}), "Flourishing Defenses — create a 1/1 green Elf Warrior creature token?"),
		},
	})
}
