package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana-Charged Dragon — Creature — Dragon {4}{R}{R}, 5/5 (Commander 2011,
// #5574):
//
//	"Flying, trample
//	 Join forces — Whenever this creature attacks or blocks, each player
//	 starting with you may pay any amount of mana. This creature gets +X/+0
//	 until end of turn, where X is the total amount of mana paid this way."
//
// One trigger for both attacking and blocking. The join forces process
// (S82OldB456JoinForces) asks each living seat in turn order from the
// Dragon's controller, then the Dragon gets +X/+0 until end of turn (last-known
// object if it has left: nothing to pump).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "87354a7b-01be-419e-b383-4c28e12a2f2b",
		Name:            "Mana-Charged Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks,
				"Mana-Charged Dragon — each player may pay any amount of mana; this creature gets +X/+0",
				func(g *game.Game, item *game.StackItem) error {
					return manaChargedDragonJoin.Start(effects.NewContext(g, item))
				}),
		},
	})
}

// manaChargedDragonJoin is the process. An on-disk key.
var manaChargedDragonJoin = effects.NewS82OldB456JoinForces("pay-unless/old-b456-mana-charged-dragon", "Mana-Charged Dragon",
	"this creature gets +X/+0 until end of turn",
	func(ctx *effects.Context, total int) error {
		if total <= 0 || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
			return nil
		}
		return effects.BoostUntilEOT{Target: ctx.Source(), Power: total, Label: "Mana-Charged Dragon — +X/+0 until end of turn"}.Apply(ctx)
	})
