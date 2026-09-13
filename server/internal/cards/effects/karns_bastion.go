package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karn's Bastion — Land:
//
//	"{T}: Add {C}.
//	 {4}, {T}: Proliferate."
//
// A colourless land that turns spare mana into counters every turn —
// the reason a superfriends or +1/+1-counter deck runs it over a
// basic. Two abilities of two different kinds on one permanent:
//
//   - "{T}: Add {C}" is a mana ability (CR 605.3, doesn't use the
//     stack) and sits at index 0 so the auto-tapper reaches for it.
//   - "{4}, {T}: Proliferate" is an ordinary activated ability that
//     DOES use the stack, so it can be responded to. Both cost
//     components tap the same permanent, so activating either locks
//     out the other for the turn — the engine validates the whole
//     cost before paying any of it.
//
// No simplification beyond the shared proliferate auto-pick.
func init() {
	Register(Spec{
		OracleID: "9fb8cd81-403a-4988-8f1c-b8eccf8abd9c",
		Name:     "Karn's Bastion",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{4}, {T}: Proliferate",
			Cost:  Plus(ManaCost("{4}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Proliferate{}.Apply(NewContext(g, item))
			},
		}},
	})
}
