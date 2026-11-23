package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guarded Heir — Creature — Human Noble {5}{W}, 1/1 (Foundations,
// #4465):
//
//	"Lifelink
//	 When this creature enters, create two 3/3 white Knight creature
//	 tokens."
//
// Maalfeld Twins' two tokens on an enters trigger instead of a dies
// one. The 3/3 white Knight is a plain token-table row; lifelink rides
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "63e79a6f-ed98-480a-a740-d44e111046f7",
		Name:            "Guarded Heir",
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Guarded Heir — create two 3/3 white Knight tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 white Knight"), N: 2})),
		},
	})
}
