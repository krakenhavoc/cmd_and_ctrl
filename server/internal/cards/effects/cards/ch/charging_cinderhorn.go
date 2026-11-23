package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Charging Cinderhorn — Creature — Elemental Ox {3}{R}, 4/2 (Commander
// 2016, #5537):
//
//	"Haste
//	 At the beginning of each player's end step, if no creatures attacked
//	 this turn, put a fury counter on this creature. Then this creature
//	 deals damage equal to the number of fury counters on it to that
//	 player."
//
// Mana Cache's every-end-step trigger with an intervening "if" (CR
// 603.4): it triggers only when no player declared an attacker this turn
// (each seat's turn tally, Chart a Course's read) and does nothing if a
// creature has attacked by the time it resolves. "That player" is the
// player whose end step it is. The counter goes on first; the damage is
// the next step, so it counts the counter just added, and a Cinderhorn
// that has left the battlefield deals damage equal to the fury counters
// it last had (CR 608.2h) without gaining one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "397e0d30-a3bf-4b33-98b6-b880f8de5cbb",
		Name:            "Charging Cinderhorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return !chargingCinderhornAnyAttack(g)
			}, "Charging Cinderhorn — put a fury counter on it, then it deals damage equal to its fury counters to that player",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						if chargingCinderhornAnyAttack(ctx.Game) || effects.SourceIsNewObject(ctx.Game, ctx.Item) || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
							return nil
						}
						return effects.AddCounter{Target: ctx.Source(), Kind: chargingCinderhornFury, N: 1}.Apply(ctx)
					}),
					effects.Instruction(chargingCinderhornDamage),
				)),
		},
	})
}

const chargingCinderhornFury = "fury"

// chargingCinderhornAnyAttack is "a creature attacked this turn": any
// player declared an attacker.
func chargingCinderhornAnyAttack(g *game.Game) bool {
	for _, p := range g.Seats {
		if p != nil && effects.B18AttackedThisTurn(g, p.ID) {
			return true
		}
	}
	return false
}

// chargingCinderhornDamage is "then this creature deals damage equal to
// the number of fury counters on it to that player".
func chargingCinderhornDamage(ctx *effects.Context) error {
	if chargingCinderhornAnyAttack(ctx.Game) {
		return nil
	}
	player := effects.TriggeringActor(ctx.Item)
	if player == uuid.Nil {
		return nil
	}
	src := ctx.Source()
	n := 0
	if effects.OnBattlefield(ctx.Game, src) && !effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		if c, ok := ctx.Game.LookupCardForEffect(src); ok {
			n = c.Counters[chargingCinderhornFury]
		}
	} else {
		n = ctx.Game.LastKnownCountersForEffect(src)[chargingCinderhornFury]
	}
	if n <= 0 {
		return nil
	}
	return effects.DealDamage{Source: src, Target: player, Amount: n}.Apply(ctx)
}
