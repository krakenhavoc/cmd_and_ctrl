package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oyobi, Who Split the Heavens — Legendary Creature — Spirit {6}{W}, 3/6:
//
//	"Flying
//	 Whenever you cast a Spirit or Arcane spell, create a 3/3 white Spirit
//	 creature token with flying."
//
// Callow Jushi's Spirit-or-Arcane cast trigger; it fires on casting, so a
// countered spell still makes the token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "df5c1f34-94e4-4f89-b1af-943c430862d0",
		Name:            "Oyobi, Who Split the Heavens",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Subtype("Spirit"), Subtype("Arcane")),
				"Oyobi — create a 3/3 white Spirit creature token with flying",
				Do(CreateToken{Template: TokenCard("3/3 white Spirit with flying"), N: 1})),
		},
	})
}
