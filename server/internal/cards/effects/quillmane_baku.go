package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quillmane Baku — Creature — Spirit {4}{U}, 3/3:
//
//	"Whenever you cast a Spirit or Arcane spell, you may put a ki
//	 counter on this creature.
//	 {1}, {T}, Remove X ki counters from this creature: Return target
//	 creature with mana value X or less to its owner's hand."
//
// ADR 0109 §9 (#1842). X is the number of ki counters the cost
// removes, announced with the activation (CR 602.2b, 107.3a) before
// the target is chosen, so the target clause is a mana-value bound
// whose input is the counters removed
// (WithManaValueAtMostX().BoundByTheCountersRemoved()), re-checked as
// the ability resolves (CR 608.2b). X may be 0, which returns a
// creature with mana value 0 — a token, most often.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "736168bb-0a5c-4b29-9f82-1319cd112873",
		Name:         "Quillmane Baku",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WheneverYouCast(Or(Subtype("Spirit"), Subtype("Arcane")),
				"Quillmane Baku — put a ki counter on this creature", putACounterOnThis("ki")),
				"Quillmane Baku — put a ki counter on it?"),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}, Remove X ki counters from this creature: Return target creature with mana value X or less to its owner's hand.",
			Cost:  Plus(ManaCost("{1}"), TapCost(), RemoveCountersXFromThis("ki", 0)),
			Targets: TargetCreature("target creature with mana value X or less").
				WithManaValueAtMostX().BoundByTheCountersRemoved(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return BounceToHand{Target: id}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
