package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wandertale Mentor — Creature — Raccoon Bard {R}{G}, 2/2 (Bloomburrow,
// #4467):
//
//	"Whenever you expend 4, put a +1/+1 counter on this creature. (You
//	 expend 4 as you spend your fourth total mana to cast spells during
//	 a turn.)
//	 {T}: Add {R} or {G}."
//
// Expend is WheneverYouExpend (expend.go), the turn's tally of mana
// spent on spells. The counter goes on this object only, so a Mentor that
// left before the trigger resolved gets nothing (CounterOnThis). The mana
// ability is a plain choice of {R} or {G}, with summoning sickness
// applying as for any creature's tap ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "9d14c7bd-7802-401c-b02b-c7f3c51b007f",
		Name:           "Wandertale Mentor",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouExpend(4, "Wandertale Mentor — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
