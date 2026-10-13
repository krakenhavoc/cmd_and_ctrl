package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Callow Jushi // Jaraku the Interloper — {1}{U}{U} Creature — Human
// Wizard 2/2 // Legendary Creature — Spirit 3/4, a flip card (#3216,
// ADR 0151):
//
//	Top:     "Whenever you cast a Spirit or Arcane spell, you may put a
//	          ki counter on this creature.
//	          At the beginning of the end step, if there are two or more
//	          ki counters on this creature, you may flip it."
//	Flipped: "Remove a ki counter from Jaraku: Counter target spell
//	          unless its controller pays {2}."
//
// The counter-unless is CounterUnlessPaid, whose prompt goes to the
// SPELL's controller (Daze's shape) and holds the stack while it is
// open. Jaraku is blue and {1}{U}{U} (CR 710.1c).
//
// No simplification.
const callowJushiOracleID = "7ceaddba-6301-45f8-86a1-ad4c6e6d3297"

func init() {
	Register(Spec{
		OracleID:     callowJushiOracleID,
		Name:         "Callow Jushi",
		Completeness: CompletenessFull,
		Triggered:    kiFlipTriggers("Callow Jushi"),
	})
	Register(Spec{
		OracleID:     callowJushiOracleID + "#1",
		Name:         "Jaraku the Interloper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Remove a ki counter from Jaraku: Counter target spell unless its controller pays {2}.",
			Cost:    removeAKiCounter(),
			Targets: TargetSpell("target spell"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return CounterUnlessPaid{
						StackID:  id,
						Cost:     "{2}",
						Question: "Jaraku the Interloper — pay {2} or your spell is countered",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
