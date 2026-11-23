package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dread Presence — Creature — Nightmare {3}{B}, 3/3 (Jumpstart 2022, #5948):
//
//	"Whenever a Swamp you control enters, choose one —
//	 • You draw a card and you lose 1 life.
//	 • This creature deals 2 damage to any target and you gain 2 life."
//
// Koth of the Homestead's Swamp-enters test (a land you control with the
// Swamp subtype, read as it entered) and Silverback Elder's real "choose
// one" on the trigger: the mode, and the target of the second, are
// announced as it goes on the stack (CR 603.3c). The damage is the
// Presence's, from its last known information if it has gone
// (CR 608.2h); the life is gained only when the mode's target is still
// legal, since a mode whose only target is gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	trig := effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
		return ok && c.IsLand() && c.HasSubtype(source.TextSubtype("Swamp"))
	}, "Dread Presence — a Swamp you control entered: choose one", func(*game.Game, *game.StackItem) error { return nil })
	trig.Modes = effects.ChooseOne(
		effects.ModeDoing("You draw a card and you lose 1 life.", nil,
			effects.DrawACardAndLoseOneLifeMode),
		effects.ModeWithPurpose(effects.ModeDoing("This creature deals 2 damage to any target and you gain 2 life.", effects.TargetAny(),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
					return err
				}
				return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
			}), effects.ForTargets(effects.DamageToTarget(0, 2))),
	)
	effects.Register(effects.Spec{
		OracleID:     "1690a782-9e54-4231-9a72-abae5ed7050c",
		Name:         "Dread Presence",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trig},
	})
}
