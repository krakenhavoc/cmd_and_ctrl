package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mysterio's Phantasm — Creature — Illusion Villain {1}{U}, 1/3
// (Marvel's Spider-Man, #4460):
//
//	"Flying, vigilance
//	 Whenever this creature attacks, mill a card."
//
// Two printed keywords and one attack trigger on the creature itself
// (ThisAttacked); the controller mills one card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "99419378-8a23-44f3-aebe-e62d6e20e877",
		Name:            "Mysterio's Phantasm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Mysterio's Phantasm — mill a card",
				Do(MillCards{N: 1})),
		},
	})
}
