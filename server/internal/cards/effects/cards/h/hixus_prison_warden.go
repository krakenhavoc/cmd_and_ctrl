package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hixus, Prison Warden — Legendary Creature — Human Soldier {3}{W}{W},
// 4/4 (Magic Origins):
//
//	"Flash
//	 Whenever a creature deals combat damage to you, if Hixus entered
//	 this turn, exile that creature until Hixus leaves the battlefield."
//
// Witch-king of Angmar's "a creature deals combat damage to you" watch
// with an intervening "if" (CR 603.4): Hixus must have entered this turn
// when the damage is dealt and again as the trigger resolves. The
// creature is the damage's source, exiled the Oblivion Ring way
// (CR 610.3): it returns when Hixus leaves, and if Hixus has left before
// this resolves nothing is exiled. A creature that is no longer the same
// permanent on the battlefield is not exiled (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "02bff648-fbad-4d61-ab09-93993a5fda58",
		Name:            "Hixus, Prison Warden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.CreatureDealtCombatDamageToYou(ev, source, ch, g) && g.EnteredThisTurn(source.InstanceID)
			}, "Hixus, Prison Warden — exile that creature until Hixus leaves the battlefield", hixusExile),
		},
	})
}

func hixusExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Source == uuid.Nil {
		return nil
	}
	if !g.EnteredThisTurn(item.SourceCardID) {
		return nil
	}
	id := item.Trigger.Event.Source
	if !effects.OnBattlefield(g, id) {
		return nil
	}
	return effects.ExileUntil{Target: id, ThisLeaves: true,
		Label: "Hixus, Prison Warden — the exiled creature returns when Hixus leaves the battlefield"}.Apply(effects.NewContext(g, item))
}
