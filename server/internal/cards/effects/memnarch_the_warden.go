package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memnarch, the Warden — Legendary Artifact Creature — Wizard {10},
// 8/9:
//
//	"Indestructible
//	 When Memnarch enters, create two 1/1 colorless Myr artifact
//	 creature tokens.
//	 Whenever Memnarch attacks, draw a card for each artifact you
//	 control."
//
// The artifacts are counted as the attack trigger resolves, so the two
// Myr (and Memnarch himself) count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f2630b02-fb48-4167-8a47-8284f7d4c484",
		Name:            "Memnarch, the Warden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Memnarch, the Warden — create two 1/1 Myr artifact creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 colorless Myr artifact"), N: 2})),
			WheneverThisAttacks("Memnarch, the Warden — draw a card for each artifact you control",
				func(g *game.Game, item *game.StackItem) error {
					n := countControlled(g, item.Controller, game.Card.IsArtifact)
					return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
