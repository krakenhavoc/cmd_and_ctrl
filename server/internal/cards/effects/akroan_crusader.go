package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d981d85c-e97b-4fe7-b415-6e73a3e7e35c",
		Name:         "Akroan Crusader",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			heroic("Akroan Crusader — create a 1/1 red Soldier creature token with haste (heroic)",
				Do(CreateToken{Template: TokenCard("1/1 red Soldier with haste"), N: 1})),
		},
	})
}
