package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chimney Rabble — Creature — Phyrexian Goblin Warrior {3}{R}, 3/3:
//
//	"Haste
//	 When this creature enters, create a 1/1 red Phyrexian Goblin
//	 creature token."
//
// The token is the shared "1/1 red Phyrexian Goblin" row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6b499d5-0e39-46ce-9de9-f64a2cc2cb32",
		Name:            "Chimney Rabble",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Chimney Rabble — create a 1/1 red Phyrexian Goblin", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Phyrexian Goblin"), N: 1})),
		},
	})
}
