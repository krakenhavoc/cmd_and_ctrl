package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse of Wizardry — Enchantment {2}{B}{B} (Rise of the Eldrazi, S82 OLD
// batch 481, #5581):
//
//	"As this enchantment enters, choose a color.
//	 Whenever a player casts a spell of the chosen color, that player loses
//	 1 life."
//
// The colour is the stored "as this enters" choice (ChooseColorAsEnters);
// the trigger reads it off the Curse. Any player's cast triggers it, and
// the caster (the cast event's actor) loses the life: life loss, not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0d78210-ba85-439f-bf1f-92ef80b6030d",
		Name:         "Curse of Wizardry",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForHarm, "Curse of Wizardry"),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || source.ChosenColor == "" {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.HasColor(source.ChosenColor)
			}, "Curse of Wizardry — that player loses 1 life",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.GainLife{Player: ctx.Item.Trigger.Event.Actor, Amount: -1}.Apply(ctx)
				}))),
		},
	})
}
