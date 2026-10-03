package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kurbis, Harvest Celebrant — Legendary Creature — Treefolk {X}{G}{G},
// 0/0:
//
//	"Kurbis enters with a number of +1/+1 counters on it equal to the
//	 amount of mana spent to cast it.
//	 Remove a +1/+1 counter from Kurbis: Prevent all damage that would be
//	 dealt this turn to another target creature with a +1/+1 counter on
//	 it."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the entry counters are a CR 614.1c
// clause over the cast (game.CastCounts.ManaSpent, #1735), and the
// ability is the not-one-use shield pinned to the target. Whether the
// target still has a +1/+1 counter is checked again as the ability
// resolves (CR 608.2b); the shield then lasts the turn.
//
// Declared caveat, the one every mana-spent reader carries (Mockingbird,
// Abby): with strict mana off the engine does not know what was spent,
// and answers zero, the weaker direction (ADR 0068 §3).
func init() {
	Register(Spec{
		OracleID:     "991fab9c-8554-4e64-a933-6cc2ffa4edb1",
		Name:         "Kurbis, Harvest Celebrant",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"With strict mana off, the game doesn't track how much mana you spent, so Kurbis enters with no counters — turn strict mana on for it to count.",
		},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{{
			Kind:  game.CounterPlusOne,
			Count: func(cast game.CastCounts) int { return cast.ManaSpent },
		}},
		Activated: []ActivatedAbility{shieldTargetRow(
			"Remove a +1/+1 counter from Kurbis: Prevent all damage that would be dealt this turn to another target creature with a +1/+1 counter on it.",
			RemoveCountersFromThis(game.CounterPlusOne, 1),
			Another(TargetCreature("another target creature with a +1/+1 counter on it", CreatureWithCounter(game.CounterPlusOne))))},
	})
}
