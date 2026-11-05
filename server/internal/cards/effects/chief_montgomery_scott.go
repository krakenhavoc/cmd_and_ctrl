package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chief Montgomery Scott — Legendary Creature — Human Artificer {2}{G}{U},
// 3/3 (Star Trek Commander, #4974):
//
//	"Strange new worlds — {6}, {T}: You may put a creature or Spacecraft
//	 card from your hand onto the battlefield. It enters with X additional
//	 +1/+1 counters on it if it's a creature and X additional charge
//	 counters on it if it's a Spacecraft, where X is the number of
//	 differently named lands you control."
//
// "Strange new worlds" is an ability word with no rules meaning. The {T}
// keeps summoning sickness in force (CR 302.6). X is counted as the
// ability resolves, each land you control counted once per name. The put
// is optional and is never a cast. The counters are the entry's own (CR
// 614.1c), so a station threshold or an "enters with" reader sees them as
// the permanent arrives, and which of the two clauses applies is read from
// the permanent as it would exist on the battlefield (CR 614.12): a
// Spacecraft gets charge counters, a creature +1/+1 counters, and a
// Spacecraft that would enter as a creature gets both. "Additional" means
// on top of any counters it would enter with anyway.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07f9c121-8a12-4f28-ae89-ec71f073b9ba",
		Name:         trekB39ScottName,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Strange new worlds — {6}, {T}: You may put a creature or Spacecraft card from your hand onto the battlefield. It enters with X additional +1/+1 counters on it if it's a creature and X additional charge counters on it if it's a Spacecraft, where X is the number of differently named lands you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{6}"), TapCost()),
			Effect:  trekB39ScottPut,
		}},
	})
}
