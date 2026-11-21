package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pack Rat — Creature — Rat {1}{B}, */* (Ravnica Remastered, #5948):
//
//	"Pack Rat's power and toughness are each equal to the number of Rats you
//	 control.
//	 {2}{B}, Discard a card: Create a token that's a copy of this creature."
//
// A layer 7a characteristic-defining ability that counts the Rats its
// controller controls as the layers are built (itself included, and every
// Rat token, changelings by their effective subtypes). The discard is a
// cost chosen as the ability is activated (CR 602.2b). The token is a copy
// of the Rat as it last existed if it has left by resolution (CR 608.2h),
// so it carries the same ability and the same counting static.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5632c2df-71e6-41b7-b8f5-c683812cc0e2",
		Name:         "Pack Rat",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Pack Rat's power and toughness are each equal to the number of Rats you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := len(permanentsControlledByMatching(g, source.Controller, Subtype("Rat")))
				c.Power, c.Toughness = n, n
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Discard a card: Create a token that's a copy of this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{B}"), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CreateTokenCopy{Controller: ctx.Controller(), Copy: item.SourceCardID, N: 1}.Apply(ctx)
			},
		}},
	})
}
