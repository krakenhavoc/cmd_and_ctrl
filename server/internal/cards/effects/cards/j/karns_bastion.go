package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// No simplification: the player chooses what to proliferate (#2525).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fb8cd81-403a-4988-8f1c-b8eccf8abd9c",
		Name:         "Karn's Bastion",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Proliferate",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Proliferate{}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
