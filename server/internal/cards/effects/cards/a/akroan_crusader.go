package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akroan Crusader — Creature — Human Soldier {R}, 1/1 (Foundations
// Jumpstart, #5929):
//
//	"Heroic — Whenever you cast a spell that targets this creature, create
//	 a 1/1 red Soldier creature token with haste."
//
// The shared heroic trigger (CR 207.2c), with a token for its effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d981d85c-e97b-4fe7-b415-6e73a3e7e35c",
		Name:         "Akroan Crusader",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Akroan Crusader — create a 1/1 red Soldier creature token with haste (heroic)",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Soldier with haste"), N: 1})),
		},
	})
}
