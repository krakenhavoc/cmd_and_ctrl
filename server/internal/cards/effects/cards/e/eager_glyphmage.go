package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eager Glyphmage — Creature — Cat Cleric {3}{W}, 3/3 (Secrets of
// Strixhaven, #4456):
//
//	"When this creature enters, create a 1/1 white and black Inkling
//	 creature token with flying."
//
// An enters trigger that makes one token, the 1/1 white-and-black Inkling
// flier the set's other Inkling makers use (tokens_table.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "070aa6c7-4741-47de-b68a-629aa4201ac0",
		Name:         "Eager Glyphmage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Eager Glyphmage — create a 1/1 white and black Inkling with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white and black Inkling with flying"), N: 1})),
		},
	})
}
