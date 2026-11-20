package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merciless Eternal — Creature — Zombie Cleric {2}{B}, 2/2
// (Hour of Devastation, #5534):
//
//	"Afflict 2 (Whenever this creature becomes blocked, defending player
//	 loses 2 life.)
//	 {2}{B}, Discard a card: This creature gets +2/+2 until end of turn."
//
// Afflict is the shared Afflict row (CR 702.130a): when this creature
// becomes blocked, the defending player loses the life, which is life
// loss and not damage.
// The pump's discard is a cost, paid as it is activated (CR 602.2b), on
// the shared thisGetsUntilEndOfTurn body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55c419e9-43f2-4d2d-8c13-057a48d47b53",
		Name:         "Merciless Eternal",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Afflict(2, "Merciless Eternal")},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Discard a card: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    Plus(ManaCost("{2}{B}"), DiscardACard()),
			Effect:  thisGetsUntilEndOfTurn(2, 2, "Merciless Eternal — +2/+2 until end of turn"),
		}},
	})
}
