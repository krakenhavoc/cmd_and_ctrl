package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Envoy of Okinec Ahau — Creature — Cat Advisor {2}{W}, 3/3 (The Lost
// Caverns of Ixalan, #4470):
//
//	"{4}{W}: Create a 1/1 colorless Gnome artifact creature token."
//
// An activated ability with no tap symbol, so it may be activated any
// time its controller has priority, as many times as they can pay. The
// token is the table's 1/1 colorless Gnome artifact creature, Oltec Cloud
// Guard's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebc70734-bf1a-4754-8094-bcbb4db91f30",
		Name:         "Envoy of Okinec Ahau",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Create a 1/1 colorless Gnome artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Cost:    ManaCost("{4}{W}"),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 colorless Gnome artifact"), N: 1}),
		}},
	})
}
