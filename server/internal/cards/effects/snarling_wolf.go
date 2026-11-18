package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snarling Wolf — Creature — Wolf {G}, 1/1 (Innistrad: Midnight Hunt, S82):
//
//	"{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
//
// The once-a-turn limit is the activation tally (CR 602.1b), as on Aggressive Mining.
//
// No simplification.
func init() {
	const label = "{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "2c181479-6df0-44ef-a7ce-d1d592735bc6",
		Name:         "Snarling Wolf",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:      ManaCost("{1}{G}"),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(2, 2, "Snarling Wolf — +2/+2 until end of turn"),
		}},
	})
}
