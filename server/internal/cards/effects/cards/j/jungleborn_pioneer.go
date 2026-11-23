package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jungleborn Pioneer — Creature — Merfolk Scout {2}{G}, 2/2 (Rivals of
// Ixalan, #5531):
//
//	"When this creature enters, create a 1/1 blue Merfolk creature token
//	 with hexproof. (It can't be the target of spells or abilities your
//	 opponents control.)"
//
// The token is Deeproot Pilgrimage's 1/1 blue Merfolk with hexproof from
// the token table.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60662463-6050-46c6-84ab-64e203b67b4d",
		Name:         "Jungleborn Pioneer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jungleborn Pioneer — create a 1/1 blue Merfolk with hexproof",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Merfolk with hexproof"), N: 1})),
		},
	})
}
