package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blaring Recruiter — Creature — Elf Warrior {3}{W}, 2/2:
//
//	"Partner with Blaring Captain (When this creature enters, target
//	 player may put Blaring Captain into their hand from their library,
//	 then shuffle.)
//	 {2}{W}: Create a 1/1 white Warrior creature token."
//
// Not legendary, so it can't be a commander: of CR 702.124j's two
// abilities only the entry search does anything (PartnerWith, #2142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33e9db5a-986d-4731-88c9-87de6f967db4",
		Name:         "Blaring Recruiter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Blaring Recruiter", "Blaring Captain"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}: Create a 1/1 white Warrior creature token.",
			Cost:    ManaCost("{2}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect:  Do(CreateToken{Template: TokenCard("1/1 white Warrior"), N: 1}),
		}},
	})
}
