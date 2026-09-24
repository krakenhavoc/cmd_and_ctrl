package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devoted Druid — Creature — Elf Druid {1}{G}, 0/2 (EDHREC rank
// 1602):
//
//	"{T}: Add {G}.
//	 Put a -1/-1 counter on this creature: Untap this creature."
//
// The combo mana creature, and the card that asked for a cost that
// ADDS a counter (#789). Two abilities:
//
//   - "{T}: Add {G}" is an ordinary mana ability, and a creature
//     source, so it waits out summoning sickness (CR 302.6).
//   - The untapper's whole cost is putting a -1/-1 counter on
//     itself: AbilityCost.AddCounter. It has no {T}, so it can be
//     activated while the Druid is tapped — which is the point — and
//     it uses the stack, as printed, so an opponent can respond
//     between the counter going on and the untap.
//
// Two rules the engine enforces rather than the card:
//
//   - CR 614.16: a counter-doubling replacement applies only to a
//     counter placed by an EFFECT, so the counter is NOT
//     replaceable. A Doubling Season does not make the untapper cost
//     two counters, and nothing stops the counter going on.
//   - CR 118.3: a permanent that cannot have the counter put on it
//     cannot pay, and the activation is refused before anything else
//     is spent.
//
// The Druid dies to the CR 704.5f state-based check once the -1/-1
// counters reach its toughness, which on a printed 0/2 is the second
// activation — so it untaps twice and then falls over, exactly as in
// paper. The auto-tapper never uses the untapper (an added counter is
// a resource the player never agreed to spend), but it does plan the
// {G}, once per untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb814e16-acf7-41d5-a357-1323dcc369f3",
		Name:         "Devoted Druid",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label: "Put a -1/-1 counter on this creature: Untap this creature.",
			Cost:  AddCounterToThis(game.CounterMinusOne, 1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.UntapTargetForEffect(item.SourceCardID)
			},
		}},
	})
}
