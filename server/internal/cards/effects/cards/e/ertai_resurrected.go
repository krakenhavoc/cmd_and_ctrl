package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ertai Resurrected — Legendary Creature — Phyrexian Human Wizard
// {2}{U}{B}, 3/2 (EDHREC rank 3404):
//
//	"Flash
//	 When Ertai Resurrected enters, choose up to one —
//	 • Counter target spell, activated ability, or triggered
//	   ability. Its controller draws a card.
//	 • Destroy another target creature or planeswalker. Its
//	   controller draws a card."
//
// A flash body that is a Cancel or a Murder, with a card for the
// victim either way. Flash rides PrintedKeywords, so he can be cast
// with a spell on the stack and his entry trigger goes on the stack
// above it.
//
// The modal entry trigger is a real "choose up to one" (Modes with a
// floor of zero, CR 603.3c). The counter bullet takes a spell OR an
// ability (TargetSpellOrAbility, #1211); the destroy bullet takes
// another creature or planeswalker. The victim's controller is read
// before the counter or the destruction moves anything, and draws
// whether or not an indestructible creature survives — the draw is not
// conditional on the removal, as printed. A target gone in response
// does nothing and draws nothing (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3d038f7c-95fa-4b71-8f74-b9b4dd45cde0",
		Name:            "Ertai Resurrected",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Key:       effects.B32ErtaiLabel,
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: effects.ChooseN("Choose up to one", 0, 1,
				effects.ModeDoing("Counter target spell, activated ability, or triggered ability. Its controller draws a card.",
					effects.TargetSpellOrAbility("target spell, activated ability, or triggered ability"),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						victim := ctx.Game.StackItemForEffect(t.ID)
						if victim == nil {
							return nil
						}
						controller := victim.Controller
						if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
							return err
						}
						return effects.DrawCards{Player: controller, N: 1}.Apply(ctx)
					}),
				effects.ModeDoing("Destroy another target creature or planeswalker. Its controller draws a card.",
					effects.Harmful(effects.Another(effects.TargetPermanent("another target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())))),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
							return nil
						}
						c, ok := ctx.Game.LookupCardForEffect(t.ID)
						if !ok {
							return nil
						}
						controller := c.Controller
						if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						return effects.DrawCards{Player: controller, N: 1}.Apply(ctx)
					}),
			),
		}},
	})
}
