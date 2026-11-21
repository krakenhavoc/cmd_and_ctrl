package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Risona, Asari Commander — {1}{R}{W} Legendary Creature — Human Samurai
// 3/3 (Kamigawa: Neon Dynasty, S82 NEO batch 20):
//
//	"Haste
//	 Whenever Risona deals combat damage to a player, if it doesn't have an
//	 indestructible counter on it, put an indestructible counter on it.
//	 Whenever combat damage is dealt to you, remove an indestructible
//	 counter from Risona."
//
// The first trigger has an intervening "if" (CR 603.4): it is checked as
// the damage is dealt and again as the ability resolves. An indestructible
// counter is a keyword counter (CR 122.1b). The second trigger fires once
// per simultaneous combat damage event to its controller (CR 603.2c, one
// batch) however many creatures dealt it, and removes a counter only if
// Risona has one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "44c4839f-c92a-4535-817b-1838c47fcb5e",
		Name:            "Risona, Asari Commander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondFewerCountersOnThis{Kind: game.CounterIndestructible, N: 1},
				effects.WheneverThisDealsCombatDamageToAPlayer(
					"Risona, Asari Commander — put an indestructible counter on it",
					effects.Do(effects.CounterOnThis{Kind: game.CounterIndestructible, N: 1}))),
			effects.OncePerBatch(effects.On(game.EventDealDamage, effects.CreatureDealtCombatDamageToYou,
				"Risona, Asari Commander — remove an indestructible counter from Risona",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					src := ctx.Source()
					card, ok := ctx.Game.LookupCardForEffect(src)
					if !ok || !effects.OnBattlefield(ctx.Game, src) || card.Counters[game.CounterIndestructible] < 1 {
						return nil
					}
					return effects.AddCounter{Target: src, Kind: game.CounterIndestructible, N: -1}.Apply(ctx)
				})))),
		},
	})
}
