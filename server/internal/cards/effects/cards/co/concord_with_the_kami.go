package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Concord with the Kami — Enchantment {3}{G} (Kamigawa: Neon Dynasty):
//
//	"At the beginning of your end step, choose one or more —
//	 • Put a +1/+1 counter on target creature with a counter on it.
//	 • Draw a card if you control an enchanted creature.
//	 • Create a 1/1 colorless Spirit creature token if you control an
//	   equipped creature."
//
// A modal end-step trigger (Black Market Connections' ChooseOneOrMore). The
// first bullet targets and is judged again as it resolves (CR 608.2b). The
// other two are not intervening ifs: the board is read as each bullet
// resolves (CR 603.4 does not apply), so an Equipment moved in an earlier
// bullet's wake still counts. "Enchanted" and "equipped" mean any Aura or
// Equipment attached, whoever controls it.
//
// No simplification.
func init() {
	t := effects.AtYourEndStep("Concord with the Kami — choose one or more",
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOneOrMore(
		effects.ModeDoing("Put a +1/+1 counter on target creature with a counter on it.",
			effects.TargetCreature("target creature with a counter on it", effects.B64HasAnyCounter()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				for _, tg := range ctx.ModeTargets(occ) {
					if tg.Kind == game.TargetCard && ctx.IsTargetLegal(tg) {
						return effects.AddCounter{Target: tg.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}
				}
				return nil
			}),
		effects.ModeDoing("Draw a card if you control an enchanted creature.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if !effects.S82NeoB11ControlsCreatureWithAttached(ctx.Game, item.Controller, func(a game.Card) bool { return a.IsAura() }) {
					return nil
				}
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Create a 1/1 colorless Spirit creature token if you control an equipped creature.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if !effects.S82NeoB11ControlsCreatureWithAttached(ctx.Game, item.Controller, func(a game.Card) bool { return a.HasSubtype("Equipment") }) {
					return nil
				}
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 colorless Spirit"), N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "bec04768-c573-4886-9ca1-2511abc05a14",
		Name:         "Concord with the Kami",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
