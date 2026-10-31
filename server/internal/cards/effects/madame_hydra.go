package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Madame Hydra — Legendary Creature — Human Villain {2}{B}{R}, 2/3
// (Marvel Super Heroes, #4455):
//
//	"Whenever you cast a Villain spell, create a 2/1 black Villain
//	 creature token with menace."
//
// "Villain" is a word of Madame Hydra's text, so the spell's type is
// judged through the source's changed words (CR 612.2, YouCast); the
// trigger is on the cast, so it resolves before the spell does. The
// token is the 2/1 black Villain with menace Agents of HYDRA makes.
// Madame Hydra's own cast is not seen by her (she is not on the
// battlefield yet).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d812c8e3-a475-4ad2-a93f-7208ad873676",
		Name:         "Madame Hydra",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Subtype("Villain"), "Madame Hydra — create a 2/1 black Villain creature token with menace",
				Do(CreateToken{Template: TokenCard("2/1 black Villain with menace"), N: 1})),
		},
	})
}
