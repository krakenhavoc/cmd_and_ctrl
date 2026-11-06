package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Biomechan Engineer — Creature — Insect Artificer {G}{U}, 2/2 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, create a Lander token. (It's an artifact
//	 with "{2}, {T}, Sacrifice this token: Search your library for a
//	 basic land card, put it onto the battlefield tapped, then
//	 shuffle.")
//	 {8}: Draw two cards and create a 2/2 colorless Robot artifact
//	 creature token."
//
// The Lander is Horizon Explorer's (b16LanderToken). The activated
// ability has no tap symbol and no timing restriction, so it can be
// activated at instant speed, any number of times; it draws first and
// then makes Secluded Starforge's Robot, the token waiting for any
// prompt a draw asks (CR 608.2c). Instant-speed creature tokens make
// it a blocker maker (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b395925b-2456-454e-9262-3b55c9f2799a",
		Name:         "Biomechan Engineer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Biomechan Engineer — create a Lander token",
				Do(CreateToken{Template: b16LanderToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{8}: Draw two cards and create a 2/2 colorless Robot artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Draws: 2, Tokens: 1, CreatureTokens: 1},
			Cost:    ManaCost("{8}"),
			Effect: Do(
				DrawCards{N: 2},
				CreateToken{Template: TokenCard("2/2 colorless Robot artifact"), N: 1},
			),
		}},
	})
}
