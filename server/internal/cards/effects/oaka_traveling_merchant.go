package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// O'aka, Traveling Merchant — Legendary Creature — Human Citizen {1}{U},
// 1/2:
//
//	"{T}, Remove a counter from a nonland permanent you control: Draw a
//	 card."
//
// Fain, the Broker's counter cost on any nonland permanent: "a counter"
// is any kind (RemoveCountersFrom with no kind), the activator names the
// permanent and the kind at announce, and the counter comes off then
// (#625), so a proliferate in response cannot bank it. The permanent is
// chosen, not targeted, and O'aka may pay from herself. The {T} waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed0fc294-de9b-4c64-b97d-e3a41fd2f46a",
		Name:         "O'aka, Traveling Merchant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove a counter from a nonland permanent you control: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), RemoveCountersFrom("", 1, "a nonland permanent you control", Nonland())),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
