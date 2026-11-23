package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Fool's Wisdom — Enchantment — Aura Curse {4}{B}{B}:
//
//	"Enchant player
//	 Whenever enchanted player draws a card, they lose 2 life and you
//	 gain 2 life.
//	 Madness {3}{B}"
//
// Curse of Opulence's enchant-player shape (Card.AttachedTo names the
// seat) watching EventDrawCard, which fires once per card, so a
// seven-card draw is seven triggers of 2 life each, as printed. The
// drawer is read off the event the trigger fired on, so it is the
// enchanted player who loses the life; the Curse's controller gains
// 2 whether or not the loss went through.
//
// Madness on an Aura is the ordinary madness cast: the discarded Curse
// waits in exile, and casting it for {3}{B} asks for its player target
// as any cast of it does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93b2a685-8267-4cbd-ac55-db6ff95fe98d",
		Name:         "Curse of Fool's Wisdom",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Madness:      "{3}{B}",
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
			}, "Curse of Fool's Wisdom — they lose 2 life and you gain 2 life",
				effects.Do(
					// Your life gain waits for a prompt their life loss asks (CR 608.2c, #4155).
					effects.Instruction(func(ctx *effects.Context) error {
						drawer := ctx.Item.Trigger.Event.Actor
						if ctx.Game.PlayerByIDForEffect(drawer) != nil {
							return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, drawer, -2)
						}
						return nil
					}),
					effects.GainLife{Amount: 2},
				)),
		},
	})
}
