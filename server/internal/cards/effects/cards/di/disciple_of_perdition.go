package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disciple of Perdition — Creature — Human Warlock {1}{B}, 1/3 (Jumpstart
// 2022, #5948):
//
//	"When this creature dies, choose one. If you have exactly 13 life, you
//	 may choose both instead.
//	 • You draw a card and you lose 1 life.
//	 • Exile target opponent's graveyard. That player loses 1 life."
//
// One dies trigger with a real "choose one" (CR 700.2): the mode, and the
// opponent of the second, are announced as it goes on the stack
// (CR 603.3c). The "exactly 13 life" count is Molten Collapse's conditional
// "you may choose both instead" (OrUpToIf), read for the controller as the
// modes are chosen (CR 700.2a), so one bullet is still a legal answer. With
// both chosen the bullets run in printed order (CR 608.2c): the draw and
// the life loss come first, so an exactly-13 controller who loses 1 life
// has 12 afterwards. The opponent loses 1 life only when still a legal
// target as the trigger resolves.
//
// No simplification.
func init() {
	trig := effects.WhenThisDies("Disciple of Perdition — choose one: you draw a card and you lose 1 life; or exile target opponent's graveyard and that player loses 1 life",
		func(*game.Game, *game.StackItem) error { return nil })
	trig.Modes = effects.ChooseOne(
		effects.ModeDoing("You draw a card and you lose 1 life.", nil,
			effects.DrawACardAndLoseOneLifeMode),
		effects.ModeDoing("Exile target opponent's graveyard. That player loses 1 life.",
			effects.TargetPlayer("target opponent", effects.Opponent()),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				if err := effects.ExileGraveyardForEffect(ctx.Game, item, t.ID); err != nil {
					return err
				}
				return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -1)
			}),
	).OrUpToIf(2, disciplePerditionExactly13Life)
	effects.Register(effects.Spec{
		OracleID:     "9ac62ff0-6d13-4091-ba84-fbfbe742a4f4",
		Name:         "Disciple of Perdition",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trig},
	})
}

// disciplePerditionExactly13Life is "If you have exactly 13 life": the
// chooser's life total as the modes are chosen.
var disciplePerditionExactly13Life = game.ModeCondition("you-have-exactly-13-life", func(g *game.Game, chooser uuid.UUID) bool {
	p := g.PlayerByIDForEffect(chooser)
	return p != nil && p.Life == 13
})
