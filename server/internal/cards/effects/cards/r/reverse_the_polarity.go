package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reverse the Polarity — Instant {1}{U}{U}:
//
//	"Choose one —
//	 • Counter all other spells.
//	 • Switch each creature's power and toughness until end of turn.
//	 • Creatures can't be blocked this turn."
//
// The counter sweep is CounterAllMatching over every spell on the stack
// but this one, the set read once as it resolves (CR 608.2). The switch
// is layer 7d (CR 613.4d, #3996) over the creatures on the battlefield
// as it resolves (CR 611.2c: switching changes a characteristic). "Can't
// be blocked" changes no characteristic, so it reaches a creature that
// enters later this turn too (game.ScopeCreatures, read live).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4499d8cd-d090-4acc-a1d9-3d9d9b6ef768",
		Name:         "Reverse the Polarity",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Counter all other spells.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					self := item.SourceCardID
					return effects.CounterAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.InstanceID != self
					}}.Apply(ctx)
				}),
			effects.ModeDoing("Switch each creature's power and toughness until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.SwitchPTUntilEOT{Match: effects.Creature(), Label: "Reverse the Polarity — switch each creature's power and toughness"}.Apply(ctx)
				}),
			effects.ModeDoing("Creatures can't be blocked this turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.RestrictUntilEOT{
						Scope:        game.ScopeCreatures,
						Restrictions: game.CantBeBlocked,
						Label:        "Reverse the Polarity — creatures can't be blocked this turn",
					}.Apply(ctx)
				}),
		),
	})
}
