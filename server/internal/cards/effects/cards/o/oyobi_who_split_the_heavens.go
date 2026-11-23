package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "df5c1f34-94e4-4f89-b1af-943c430862d0",
		Name:            "Oyobi, Who Split the Heavens",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Subtype("Spirit"), effects.Subtype("Arcane")),
				"Oyobi — create a 3/3 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 white Spirit with flying"), N: 1})),
		},
	})
}
