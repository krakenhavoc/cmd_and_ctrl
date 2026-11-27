package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alaborn Zealot — Creature — Human Soldier {W}, 1/1 (Portal Second Age, S82
// older sets batch 733, #5625):
//
//	"When this creature blocks a creature, destroy both creatures."
//
// One trigger per attacker blocked (CR 509.3b). "That creature" is the
// attacker as it was when blocked, so one that left and came back is a new
// object and is untouched (CR 400.7); the Zealot is destroyed as the source,
// by last-known information if it is gone.
//
// No simplification.
func init() {
	const label = "Alaborn Zealot — destroy it and the creature it blocks"
	t := effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return effects.B28SelfBlocked(ev, source)
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		attacker := item.Params.Object
		if attacker.ID != uuid.Nil && effects.OnBattlefield(g, attacker.ID) && g.ObjectEpochForEffect(attacker.ID) == attacker.Epoch {
			if err := (effects.DestroyTarget{Target: attacker.ID}).Apply(ctx); err != nil {
				return err
			}
		}
		if effects.OnBattlefield(g, item.SourceCardID) {
			return effects.DestroyTarget{Target: item.SourceCardID}.Apply(ctx)
		}
		return nil
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.Target, Epoch: g.ObjectEpochForEffect(ev.Target)}
		return item
	}
	effects.Register(effects.Spec{
		OracleID:     "0d0d97dd-2653-4d37-b475-ac3b50c5ee53",
		Name:         "Alaborn Zealot",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
