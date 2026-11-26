package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fable of Wolf and Owl — Enchantment {3}{G/U}{G/U}{G/U}:
//
//	"Whenever you cast a green spell, you may create a 2/2 green Wolf creature
//	 token.
//	 Whenever you cast a blue spell, you may create a 1/1 blue Bird creature
//	 token with flying."
//
// Two triggers, one per printed ability. Each reads the spell's colours as it
// was cast (CR 601.2i), so a green-blue spell triggers both, and each is a
// "you may" asked as the trigger goes on the stack (CR 603.5).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24fe8bc3-7e3a-435b-a232-7f80787aba1f",
		Name:         "Fable of Wolf and Owl",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.OfColor("G"), "Fable of Wolf and Owl — you may create a 2/2 green Wolf creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1})),
				"Fable of Wolf and Owl — create a 2/2 green Wolf creature token?"),
			effects.Optional(effects.WheneverYouCast(effects.OfColor("U"), "Fable of Wolf and Owl — you may create a 1/1 blue Bird creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Bird with flying"), N: 1})),
				"Fable of Wolf and Owl — create a 1/1 blue Bird creature token with flying?"),
		},
	})
}
