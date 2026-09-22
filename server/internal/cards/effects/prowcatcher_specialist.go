package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prowcatcher Specialist — Creature — Goblin Warrior {1}{R}, 2/2:
//
//	"Haste
//	 Exhaust — {3}{R}: Put two +1/+1 counters on this creature.
//	 (Activate each exhaust ability only once.)"
//
// The smallest exhaust card printed, and the catalog's proof that the
// keyword is one bit (#1181). Everything the reminder text promises is
// in `Exhaust: true`: the engine keys an activation record by (object,
// this label), refuses the second activation with
// ErrAbilityExhausted, drops the move from the legal enumerator and
// greys the row on the wire. There is no per-card state, no condition
// closure and nothing for the card file to forget.
//
// What the keyword is NOT: a once-each-turn clause. The record has a
// game-lifetime scope, so the counters do not come back on the next
// upkeep. What refreshes it is a new OBJECT (CR 400.7) — flicker the
// Goblin and the returning creature may exhaust again, because it is
// not the permanent that spent it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "594132ac-32a7-41d5-b7f0-3692bbed7f6f",
		Name:            "Prowcatcher Specialist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {3}{R}: Put two +1/+1 counters on this creature.",
			Exhaust: true,
			Cost:    ManaCost("{3}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AddCounter{
					Target: item.SourceCardID,
					Kind:   game.CounterPlusOne,
					N:      2,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
