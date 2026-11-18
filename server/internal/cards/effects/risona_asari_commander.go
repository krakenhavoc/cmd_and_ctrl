package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "44c4839f-c92a-4535-817b-1838c47fcb5e",
		Name:            "Risona, Asari Commander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondFewerCountersOnThis{Kind: game.CounterIndestructible, N: 1},
				WheneverThisDealsCombatDamageToAPlayer(
					"Risona, Asari Commander — put an indestructible counter on it",
					Do(CounterOnThis{Kind: game.CounterIndestructible, N: 1}))),
			OncePerBatch(On(game.EventDealDamage, creatureDealtCombatDamageToYou,
				"Risona, Asari Commander — remove an indestructible counter from Risona",
				Do(Instruction(func(ctx *Context) error {
					src := ctx.Source()
					card, ok := ctx.Game.LookupCardForEffect(src)
					if !ok || !onBattlefield(ctx.Game, src) || card.Counters[game.CounterIndestructible] < 1 {
						return nil
					}
					return AddCounter{Target: src, Kind: game.CounterIndestructible, N: -1}.Apply(ctx)
				})))),
		},
	})
}
