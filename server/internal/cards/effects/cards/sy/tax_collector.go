package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tax Collector — Creature — Human Advisor {1}{W}, 2/2 (Assassin's Creed,
// #5948):
//
//	"When this creature enters, choose one —
//	 • Tax — Until your next turn, spells your opponents cast cost {1} more to
//	   cast.
//	 • Arrest — Detain target creature an opponent controls. (Until your next
//	   turn, that creature can't attack or block and its activated abilities
//	   can't be activated.)"
//
// The mode is chosen as the trigger goes on the stack (CR 700.2); only Arrest
// has a target. Tax is a scoped record (game.ModSpellsCostMoreForOpponents,
// #5948) read by the cast pricer: every player but Tax Collector's controller
// pays {1} more for each spell until that player's next turn begins, and the
// record outlives Tax Collector. Arrest is the detain keyword action (CR 701.35,
// Azorius Arrester's Detain over the target as it resolves).
//
// No simplification.
func init() {
	const label = "Tax Collector — choose one: until your next turn, spells your opponents cast cost {1} more; or detain target creature an opponent controls"
	t := effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("Tax — Until your next turn, spells your opponents cast cost {1} more to cast.",
			nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeGame, ctx.Controller(),
					[]game.Mod{game.SpellsCostMoreForOpponentsMod(1)},
					effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
					"Tax Collector — spells your opponents cast cost {1} more to cast")
				return nil
			}),
		effects.ModeDoing("Arrest — Detain target creature an opponent controls.",
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				target, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.Detain{
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == target.ID },
					Label: "Tax Collector — detain",
				}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "66837faa-fbde-456d-b00f-eb17fbdd04aa",
		Name:         "Tax Collector",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
