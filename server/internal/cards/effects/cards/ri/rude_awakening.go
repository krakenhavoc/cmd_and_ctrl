package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rude Awakening — Sorcery {4}{G}:
//
//	"Choose one —
//	 • Untap all lands you control.
//	 • Until end of turn, lands you control become 2/2 creatures that
//	   are still lands.
//	 Entwine {2}{G} (Choose both if you pay the entwine cost.)"
//
// Entwine is CR 702.42 (#3866). The second bullet is Living Plane's two
// layers as a one-shot until end of turn: layer 4 adds the creature
// type and keeps every other type, layer 7b sets the base power and
// toughness to 2/2. The affected set is the lands you control as it
// resolves (CR 611.2c), so a land that enters later stays a land. The
// lands are summoning sick only if they came under your control this
// turn (CR 302.6). Entwined, the lands are untapped first (CR 702.42b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "685a7f67-6809-4386-990f-c52d1846a584",
		Name:          "Rude Awakening",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{2}{G}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Untap all lands you control.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.B16UntapAllYouControlMatching(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
				}),
			effects.ModeDoing("Until end of turn, lands you control become 2/2 creatures that are still lands.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.ScopedEffectFor{
						Match: effects.And(effects.Land(), effects.YouControl()),
						Mods: []game.Mod{
							game.AddTypesMod("Creature"),
							game.SetBasePowerMod(2),
							game.SetBaseToughnessMod(2),
						},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Rude Awakening — lands you control are 2/2 creatures",
					}.Apply(ctx)
				}),
		)),
	})
}
