package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhystic Syphon — Sorcery {3}{B}{B} (Prophecy, #5617):
//
//	"Unless target player pays {3}, that player loses 5 life and you
//	 gain 5 life."
//
// The target player is asked as the spell resolves; the table waits for
// the answer (#5793, CR 118.12). A decline is the drain, both halves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1eea34a-74ef-4d7c-bf70-c64b3487f296",
		Name:         "Rhystic Syphon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return effects.PayUnless{
					Chooser:  t.ID,
					Cost:     "{3}",
					Question: "Rhystic Syphon — pay {3}, or you lose 5 life and its controller gains 5 life?",
					Else:     rhysticSyphonDrain,
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// rhysticSyphonDrain is the decline. Keyed, an on-disk identity.
var rhysticSyphonDrain = effects.PayUnlessElse("pay-unless/rhystic-syphon-drain",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if err := (effects.TargetPlayerLosesLife{N: 5}).Apply(ctx); err != nil {
			return err
		}
		return effects.GainLife{Amount: 5}.Apply(ctx)
	})
