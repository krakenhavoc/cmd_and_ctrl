package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arashin War Beast — {5}{G}{G} Creature — Beast 6/6 (Fate Reforged):
//
//	"Whenever this creature deals combat damage to one or more blocking
//	 creatures, manifest the top card of your library."
//
// The damage event names the blocker as it stood when damage was dealt, so a
// blocker that dies to it still counts (CR 510.2). The damage is dealt at once,
// so the trigger is one per batch however many blockers were hurt
// (OncePerBatch, CR 603.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75c81ce4-ca1d-464a-bb39-e568f8ebf71f",
		Name:         "Arashin War Beast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventDealDamage, arashinDamagedABlocker,
				"Arashin War Beast — manifest the top card of your library", effects.Do(effects.Manifest{N: 1}))),
		},
	})
}

// arashinDamagedABlocker is combat damage from the source to a creature that
// is blocking.
func arashinDamagedABlocker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source != source.InstanceID {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	return ok && target.IsCreature() && len(target.BlockedAttackers()) > 0
}
