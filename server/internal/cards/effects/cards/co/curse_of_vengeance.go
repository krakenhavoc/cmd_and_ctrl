package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Vengeance — Enchantment — Aura Curse {B} (Commander 2016, #5537):
//
//	"Enchant player
//	 Whenever enchanted player casts a spell, put a spite counter on this Aura.
//	 When enchanted player loses the game, you gain X life and draw X cards,
//	 where X is the number of spite counters on this Aura."
//
// The first trigger watches the enchanted player's casts. The second watches
// EventPlayerEliminated, which is emitted while the Aura is still attached to
// the player (before CR 800.4a takes their objects); X is read off the Aura as
// it is, or as it last existed once a state-based action has put it into the
// graveyard (CR 608.2h), through PermanentForEffect (#6123's engine test pins
// both). The Aura's controller gains the life and draws the cards.
//
// No simplification.
func init() {
	enchantedPlayer := func(ev game.Event, source *game.Card) bool {
		return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
	}
	effects.Register(effects.Spec{
		OracleID:     "b9b1c644-041a-4b61-9d6d-bc73a3a95bb8",
		Name:         "Curse of Vengeance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return enchantedPlayer(ev, source)
			}, "Curse of Vengeance — put a spite counter on this Aura",
				effects.Do(effects.CounterOnThis{Kind: "spite", N: 1})),
			effects.On(game.EventPlayerEliminated, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return enchantedPlayer(ev, source)
			}, "Curse of Vengeance — you gain X life and draw X cards, where X is the number of spite counters on this Aura",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					x := 0
					if ref, ok := ctx.Game.SourceObjectForEffect(ctx.Item); ok {
						if info, ok := ctx.Game.PermanentForEffect(ref); ok {
							x = info.Counters["spite"]
						}
					}
					if x <= 0 {
						return nil
					}
					if err := (effects.GainLife{Amount: x}).Apply(ctx); err != nil {
						return err
					}
					return ctx.Game.DrawNForEffect(ctx.Controller(), x)
				}))),
		},
	})
}
