package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Truefire Captain — Creature — Human Knight {R}{R}{W}{W}, 4/3 (Ravnica
// Remastered, #5948):
//
//	"Mentor (Whenever this creature attacks, put a +1/+1 counter on target
//	 attacking creature with lesser power.)
//	 Whenever this creature is dealt damage, it deals that much damage to
//	 target player."
//
// Mentor is Blade Instructor's trigger. The damage trigger is Brash Taunter's,
// aimed at any player rather than an opponent: one trigger per damage event,
// for the amount that event dealt (a batch of simultaneous combat damage from
// several blockers is one event per source, CR 603.2c), dealt by the Captain
// as it last existed if it has since died (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3677e743-8d3d-4ec7-8bfe-21e36979a79b",
		Name:         "Truefire Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Mentor("Truefire Captain — mentor"),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.B35SelfWasDealtDamage(ev, source)
				},
				Targets:    effects.TargetPlayer("target player"),
				Key:        "Truefire Captain — it deals that much damage to target player",
				AtBatchEnd: effects.ThatMuchDamage(),
				Effect:     effects.DamagedCreatureReflectsChosenAmount,
			},
		},
	})
}
