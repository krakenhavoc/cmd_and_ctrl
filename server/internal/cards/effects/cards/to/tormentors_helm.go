package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tormentor's Helm — Artifact — Equipment {R} (Kaldheim, #5505):
//
//	"Equipped creature gets +1/+1.
//	 Whenever equipped creature becomes blocked, it deals 1 damage to
//	 defending player.
//	 Equip {1}"
//
// The damage is the equipped creature's (lifelink, deathtouch and
// protection read it), dealt to the player the blocked attacker was
// attacking, as the trigger resolves; a creature that has left deals it as
// it last existed (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f5f0988-0b1f-4f4b-bcee-cc76fa484d9e",
		Name:         "Tormentor's Helm",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.AttachedCreatureBecomesBlocked,
				"Tormentor's Helm — the blocked creature deals 1 damage to defending player",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ev := item.Trigger.Event
					return g.DealDamageToPlayerForEffect(ev.CardID, ev.Actor, 1)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
