package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const cruelSomnophageOracleID = "997bdec5-f67b-4822-a3ba-c636e2685e8a"

// Cruel Somnophage // Can't Wake Up — Creature — Nightmare {1}{B},
// */*, with an Adventure, Can't Wake Up {1}{U} (Sorcery):
//
//	"Cruel Somnophage's power and toughness are each equal to the
//	 number of creature cards in all graveyards."
//	"Target player mills four cards."
//
// The creature is Mortivore's characteristic-defining ability; the
// Adventure half is a registered "#1" entry like Bonecrusher Giant's
// Stomp, and the engine handles the exile and the later cast from it.
func init() {
	Register(Spec{
		OracleID:     cruelSomnophageOracleID,
		Name:         "Cruel Somnophage",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := b41CreatureCardsInAllGraveyards(g)
				c.Power = n
				c.Toughness = n
			},
		}},
	})

	Register(Spec{
		OracleID:     cruelSomnophageOracleID + "#1",
		Name:         "Can't Wake Up",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return MillCards{Player: victim, N: 4}.Apply(ctx)
		},
	})
}
