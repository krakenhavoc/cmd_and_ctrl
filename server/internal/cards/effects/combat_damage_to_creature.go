package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_damage_to_creature.go — "whenever this creature deals combat
// damage to a creature, <do something to> that creature" (Kaldra
// Compleat's granted trigger). Its own file so the next card that
// prints the clause appends here rather than copying the three
// pieces.

// thisDealtCombatDamageToACreature is the trigger condition: the
// source dealt combat damage (CR 510.2) to a creature. The damage
// event is emitted before the state-based actions that might kill the
// creature, so it is still on the battlefield to be asked about. A
// planeswalker or battle that took the damage is not a creature and
// does not match.
func thisDealtCombatDamageToACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source != source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok {
		return false
	}
	if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	return c.IsCreature()
}

// damagedCreatureRef names the damaged creature as the OBJECT it is
// when the trigger is built, for the trigger's Build to stamp on
// item.Params.Object. A card that is not on the battlefield and has
// not left it this turn gets an epoch no object has, so the
// resolution finds nothing to act on.
func damagedCreatureRef(g *game.Game, id uuid.UUID) game.ObjectRef {
	if ref, ok := g.PermanentRefForEffect(id); ok {
		return ref
	}
	return game.ObjectRef{ID: id, Epoch: -1}
}

// exileDamagedCreature is "exile that creature": the object named on
// item.Params.Object, and only while it is still that object on the
// battlefield. One that died to the damage is in a graveyard, and one
// that came back since is a new object (CR 400.7); neither is exiled.
func exileDamagedCreature(g *game.Game, item *game.StackItem) error {
	ref := item.Params.Object
	c, ok := g.LookupCardForEffect(ref.ID)
	if !ok || c.ObjectEpoch != ref.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(ref.ID); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return ExileTarget{Target: ref.ID}.Apply(NewContext(g, item))
}
