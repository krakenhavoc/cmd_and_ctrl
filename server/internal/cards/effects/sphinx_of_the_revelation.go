package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphinx of the Revelation — Artifact Creature — Sphinx {3}{W}{U}, 4/5:
//
//	"Flying, lifelink
//	 Whenever you gain life, you get that many {E} (energy counters).
//	 {W}{U}{U}, {T}, Pay X {E}: Draw X cards."
//
// "That many" is the life the triggering gain added (the event's
// Amount). "Pay X {E}" is PayXEnergy (ADR 0129 §2): X is announced with
// the activation (CR 107.3a), may not exceed the activator's energy
// (CR 118.3), and is what the draw reads. There is no {X} in the mana,
// so the same mana pays for any X. XMatters: a draw of zero is a no-op.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9b8f6c6a-49a1-4061-a2c4-5be612e138fa",
		Name:            "Sphinx of the Revelation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		XMatters:        true,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Sphinx of the Revelation — you get that many {E}", youGetThatManyEnergy),
		},
		Activated: []ActivatedAbility{{
			Label: "{W}{U}{U}, {T}, Pay X {E}: Draw X cards.",
			Cost:  Plus(ManaCost("{W}{U}{U}"), TapCost(), PayXEnergy()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return DrawCards{N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
