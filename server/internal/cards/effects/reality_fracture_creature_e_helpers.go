package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_e_helpers.go — slice fra-creature-e
// (tracker #2795): the pieces its creatures share that no other file
// owns. Everything is prefixed rfCreatureE.

// rfCreatureEOncePerDiscarder makes a discard trigger fire once per
// batch per discarding player — "whenever a player discards one or more
// cards" (CR 603.2c). The engine emits one EventDiscardCard per card, so
// one player's two-card discard is one trigger and two players' discards
// are two.
func rfCreatureEOncePerDiscarder(t game.TriggeredAbility) game.TriggeredAbility {
	t.OncePerBatch = true
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.Actor.String() }
	return t
}

// rfCreatureEWeakOpposingBlocker is Tetsuko's condition: the blocker is
// a creature an opponent of the source's controller controls whose
// effective power or toughness is 1 or less, on the first pair of its
// block declaration (CR 509.3a: one trigger per blocking creature).
func rfCreatureEWeakOpposingBlocker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBlock || ev.Amount > 1 || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	blocker, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !blocker.IsCreature() {
		return false
	}
	return blocker.CurrentPower() <= 1 || blocker.CurrentToughness() <= 1
}

// rfCreatureETetsukoPing deals 1 damage from the source to the
// controller of the creature that blocked.
func rfCreatureETetsukoPing(g *game.Game, item *game.StackItem) error {
	victim := item.Trigger.Event.Actor
	if victim == uuid.Nil {
		return nil
	}
	ctx := NewContext(g, item)
	return g.DealDamageToPlayerForEffect(ctx.Source(), victim, 1)
}
