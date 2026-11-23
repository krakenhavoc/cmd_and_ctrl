package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worthy Knight — Creature — Human Knight {1}{W}, 2/2:
//
//	"Whenever you cast a Knight spell, create a 1/1 white Human creature
//	 token."
//
// Academy Instructor's cast trigger with a Knight filter: it fires on
// casting, so a countered Knight still makes the token. The Knight's own
// cast is not seen (it is not on the battlefield yet).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37733e29-3ade-4bfd-89c9-efd13b548a53",
		Name:         "Worthy Knight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Subtype("Knight"), "Worthy Knight — create a 1/1 white Human creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1})),
		},
	})
}
