package game

import (
	"slices"

	"github.com/google/uuid"
)

// damage_sources_tally.go — "N or more sources you controlled dealt
// damage this turn" (Case of the Burning Masks, #4469).
//
// The count is of source OBJECTS: a permanent that deals damage several
// times without changing zones is one source, however many times it dealt
// damage, while the same card after a zone change is a new object (CR
// 400.7) and counts again. Each source is read with the controller it had
// as it dealt the damage, so a source that later dies or changes control
// still counts for the player who controlled it then (CR 608.2h).

// DamageSourceRecord is one source object that dealt damage this turn.
type DamageSourceRecord struct {
	// Controller is who controlled the source as it dealt the damage.
	Controller uuid.UUID `json:"controller"`
	// Source names the object by instance ID and ObjectEpoch.
	Source ObjectRef `json:"source"`
}

// recordDamageSourceLocked notes the source of the EventDealDamage `ev`,
// once per object. A source the engine cannot find is not recorded, which
// errs weaker. Caller must hold g.mu.
func (g *Game) recordDamageSourceLocked(ev Event) {
	if ev.Source == uuid.Nil {
		return
	}
	controller, epoch := ev.Actor, 0
	if c := g.findCardByIDLocked(ev.Source); c != nil {
		epoch = c.ObjectEpoch
		if controller == uuid.Nil {
			controller = c.Controller
		}
	}
	if controller == uuid.Nil {
		if lki := g.damageSourceLKILocked(ev.Source); lki != nil {
			controller = lki.Controller
		}
	}
	if controller == uuid.Nil {
		return
	}
	rec := DamageSourceRecord{Controller: controller, Source: ObjectRef{ID: ev.Source, Epoch: epoch}}
	if slices.Contains(g.TurnTally.DamageSources, rec) {
		return
	}
	g.TurnTally.DamageSources = append(g.TurnTally.DamageSources, rec)
}

// DamageSourcesDealtThisTurnForEffect is how many distinct source objects
// `controller` controlled as they dealt damage this turn. Caller must hold
// g.mu (an effect resolving does).
func (g *Game) DamageSourcesDealtThisTurnForEffect(controller uuid.UUID) int {
	n := 0
	for _, r := range g.TurnTally.DamageSources {
		if r.Controller == controller {
			n++
		}
	}
	return n
}
