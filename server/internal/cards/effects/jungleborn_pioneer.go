package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "60662463-6050-46c6-84ab-64e203b67b4d",
		Name:         "Jungleborn Pioneer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jungleborn Pioneer — create a 1/1 blue Merfolk with hexproof",
				Do(CreateToken{Template: TokenCard("1/1 blue Merfolk with hexproof"), N: 1})),
		},
	})
}
