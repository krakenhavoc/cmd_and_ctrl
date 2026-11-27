package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planar Chaos — Enchantment {2}{R} (Judgment):
//
//	"At the beginning of your upkeep, flip a coin. If you lose the flip,
//	 sacrifice this enchantment.
//	 Whenever a player casts a spell, that player flips a coin. If they lose
//	 the flip, counter that spell."
//
// Mana Crypt's keyed coin flip on the upkeep (CR 705.2): a lost flip
// sacrifices this enchantment only while it is still the permanent the
// trigger came from. The cast trigger watches every player's casts, the
// controller's included; the caster calls the flip, and a lost one counters
// the spell the trigger names if it is still on the stack (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Planar Chaos"
	effects.Register(effects.Spec{
		OracleID:     "f0864dbb-4d91-4282-a774-7b225e649ca8",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep(name+" — flip a coin; if you lose the flip, sacrifice this enchantment",
				func(g *game.Game, item *game.StackItem) error {
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: name + " — call the coin flip",
						ThenKey:  planarChaosUpkeepFlipped,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID != uuid.Nil
			}, name+" — that player flips a coin; if they lose the flip, counter that spell",
				func(g *game.Game, item *game.StackItem) error {
					ev := effects.NewContext(g, item).Trigger().Event
					if g.StackItemForEffect(ev.CardID) == nil {
						return nil
					}
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  ev.Actor,
						Source:   item.SourceCardID,
						Question: name + " — call the coin flip",
						ThenKey:  planarChaosCastFlipped,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// planarChaosUpkeepFlipped sacrifices the enchantment on a lost flip. An
// on-disk key: never renamed, never reused.
var planarChaosUpkeepFlipped = effects.CoinFlipThen("coin-flip/s82-old-b626-planar-chaos-sacrifice", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return nil
	}
	return effects.SacrificeThisIfStillHere(ctx)
})

// planarChaosCastFlipped counters the spell on a lost flip. An on-disk key:
// never renamed, never reused.
var planarChaosCastFlipped = effects.CoinFlipThen("coin-flip/s82-old-b626-planar-chaos-counter", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return nil
	}
	return effects.CounterTarget{StackID: ctx.Trigger().Event.CardID}.Apply(ctx)
})
