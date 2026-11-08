package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Topiary Lecturer — Creature — Elf Druid {2}{G}, 1/2 (Secrets of
// Strixhaven, #4456):
//
//	"Increment (Whenever you cast a spell, if the amount of mana you spent
//	 is greater than this creature's power or toughness, put a +1/+1 counter
//	 on this creature.)
//	 {T}: Add an amount of {G} equal to this creature's power."
//
// Increment is the keyword helper (CR 702.191). The mana ability is Marwyn,
// the Nurturer's: the amount is this creature's power as it is when the
// ability is activated, layers brought current first (b31CurrentPowerOf), so a
// counter that Increment has just put on, an anthem and a -1/-1 counter all
// count, and a power of zero adds nothing and still taps. Summoning sickness
// applies to the tap (CR 302.6). Nothing rides on the mana ability, so it
// stays a mana ability (CR 605.1a) and does not use the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5f5fe039-4be4-42ae-9fa4-4392328aa665",
		Name:            "Topiary Lecturer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{KeywordIncrement},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{Increment()},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			Label:          "Add an amount of {G} equal to this creature's power",
			ProducedAmount: tdmGreenEqualToPower,
		}},
	})
}
