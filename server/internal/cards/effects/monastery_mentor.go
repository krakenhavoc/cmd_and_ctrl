package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monastery Mentor — Creature — Human Monk {2}{W}, 2/2:
//
//	"Prowess (Whenever you cast a noncreature spell, this creature
//	 gets +1/+1 until end of turn.)
//	 Whenever you cast a noncreature spell, create a 1/1 white Monk
//	 creature token with prowess."
//
// #706's token proof card. Prowess is a canonical keyword the engine
// turns into a trigger (game/prowess.go), so the Mentor's own prowess
// is the PrintedKeywords line and the Monks' is the token row's
// Keywords — no trigger is written for either. A Monk made by one
// spell does not pump for that spell: it was not on the battlefield
// when the spell was cast. It does pump for the next one.
//
// The Mentor's two abilities trigger on the same cast and are two
// different triggers, so its controller orders them (CR 603.3b): the
// order is visible only in which resolves first, not in the result.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3979067a-9c68-443d-a85f-d9f07be880b9",
		Name:            "Monastery Mentor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Monastery Mentor — create a 1/1 Monk with prowess",
				Do(CreateToken{Template: TokenCard("1/1 white Monk with prowess"), N: 1})),
		},
	})
}
