package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cutthroat Contender — Creature — Vampire Warrior {B}, 1/1 (Streets of New
// Capenna, S82):
//
//	"Pay 1 life: This creature gets +1/+0 until end of turn. Activate only
//	 once each turn."
//
// The life is a cost, so the ability can't be announced at 0 life or less
// than 1 (CR 119.4); the limit is the per-object activation tally
// (OncePerTurnActivation), spent at announcement.
//
// No simplification.
func init() {
	const label = "Pay 1 life: This creature gets +1/+0 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "43192553-0db9-48ec-ae88-a2605a6608d8",
		Name:         "Cutthroat Contender",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      PayLife(1),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(1, 0, "Cutthroat Contender — +1/+0"),
		}},
	})
}
