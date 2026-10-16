package game

import (
	"fmt"

	"github.com/google/uuid"
)

// dealt_damage_by.go — "creatures dealt damage by this creature this
// turn" (#1946).
//
// # The record
//
// TurnTally.DamagedObjects lists every damage event dealt to a permanent
// this turn, as a pair of OBJECTS: the source and the damaged permanent,
// each an instance ID with the ObjectEpoch it had as the damage was
// dealt. The tally listener writes it from EventDealDamage, which carries
// the amount actually dealt after prevention, so damage prevented to 0
// is no damage and leaves no entry. Combat and noncombat damage alike are
// recorded, and the turn boundary clears it with the rest of the tally.
//
// Object identity is the point (CR 400.7): a creature that was dealt
// damage, flickered and came back is a new object nothing has damaged,
// and a source that left and came back is a new object that has damaged
// nothing. The record outlives both, so a reader asking about an object
// that has just died (a dies trigger, an "exile it instead" read as it
// would die) still finds it.
//
// # The readers
//
//   - ObjectDealtDamageToObjectThisTurn: by refs, for a reader that holds
//     both objects.
//   - PermanentDealtDamageToPermanentThisTurn: by card, each read as the
//     permanent object it is now or most recently was this turn
//     (PermanentRefForEffect) — Zurgo Helmsmasher's and Bushi Tenderfoot's
//     dies triggers, Kumano's "exile it instead".
//
// # The rules built on it
//
//   - ScopeDealtDamageBy (Runesword): a ScopedEffect whose mods name the
//     dealer in Objects[0], carrying cantBeRegenerated or exileIfWouldDie.
//     Its set is read live, so it reaches the creatures the dealer damages
//     after the ability resolved; DamageSince narrows it to the damage
//     dealt from a moment on.
//   - CardDef.DealtDamageCantBeRegenerated (Bone Shaman's granted
//     ability): a static of the permanent, read live under its
//     CatalogAbilityKey, so it stops the moment the permanent leaves or
//     loses the ability.
//
// Both reach regeneration through the one gate, regenerationRefusedLocked
// (CR 701.19c), and the exile through the scoped replacement's AppliesTo.

// ObjectDamageRecord is one damage event dealt to a permanent (#1946):
// the source object and the damaged object as they were when it was
// dealt. Combat marks combat damage (CR 510).
type ObjectDamageRecord struct {
	Source ObjectRef `json:"source"`
	Victim ObjectRef `json:"victim"`
	Combat bool      `json:"combat,omitempty"`
}

// recordDamagedObjectLocked notes the damage event `ev`, dealt to a
// permanent. A source the engine cannot find anywhere, or a target that
// is not on the battlefield, is not recorded, which errs weaker: nothing
// can count it. Caller must hold g.mu.
func (g *Game) recordDamagedObjectLocked(ev Event) {
	if ev.Amount <= 0 || ev.Source == uuid.Nil {
		return
	}
	victim := findBattlefieldCard(g, ev.Target)
	if victim == nil {
		return
	}
	src := g.findCardByIDLocked(ev.Source)
	if src == nil {
		return
	}
	g.TurnTally.DamagedObjects = append(g.TurnTally.DamagedObjects, ObjectDamageRecord{
		Source: ObjectRef{ID: src.InstanceID, Epoch: src.ObjectEpoch},
		Victim: ObjectRef{ID: victim.InstanceID, Epoch: victim.ObjectEpoch},
		Combat: ev.Combat,
	})
}

// objectDealtDamageSinceLocked reports whether `source` dealt damage to
// `victim` this turn in a record at index `since` or later.
func (g *Game) objectDealtDamageSinceLocked(source, victim ObjectRef, since int) bool {
	if source.ID == uuid.Nil || victim.ID == uuid.Nil {
		return false
	}
	recs := g.TurnTally.DamagedObjects
	if since < 0 {
		since = 0
	}
	for i := since; i < len(recs); i++ {
		if recs[i].Source == source && recs[i].Victim == victim {
			return true
		}
	}
	return false
}

// ObjectDealtDamageToObjectThisTurn reports whether the object `source`
// dealt damage to the permanent object `victim` this turn, combat or
// not. Either ref naming a different object of the same card — one that
// left and came back — answers false (CR 400.7).
//
// Caller must hold g.mu.
func (g *Game) ObjectDealtDamageToObjectThisTurn(source, victim ObjectRef) bool {
	return g.objectDealtDamageSinceLocked(source, victim, 0)
}

// PermanentDealtDamageToPermanentThisTurn is "<victim> was dealt damage
// by <source> this turn", each card read as the permanent object it is
// now or, once it has left, the one it most recently was this turn
// (PermanentRefForEffect): "whenever a creature dealt damage by Zurgo
// this turn dies", asked as the creature dies, and Kumano's "if a
// creature dealt damage by Kumano this turn would die". A card that is
// neither on the battlefield nor departed this turn answers false.
//
// Caller must hold g.mu.
func (g *Game) PermanentDealtDamageToPermanentThisTurn(sourceID, victimID uuid.UUID) bool {
	if len(g.TurnTally.DamagedObjects) == 0 {
		return false
	}
	source, ok := g.PermanentRefForEffect(sourceID)
	if !ok {
		return false
	}
	victim, ok := g.PermanentRefForEffect(victimID)
	if !ok {
		return false
	}
	return g.ObjectDealtDamageToObjectThisTurn(source, victim)
}

// DamageRecordMarkForEffect is the current length of the turn's damage
// record: a Mod.DamageSince that counts only the damage dealt from now
// on. Caller must hold g.mu.
func (g *Game) DamageRecordMarkForEffect() int {
	return len(g.TurnTally.DamagedObjects)
}

// ---------------------------------------------------------------
// Runesword's records: ScopeDealtDamageBy
// ---------------------------------------------------------------

// CantBeRegeneratedIfDealtDamageByForEffect is "if <dealer> deals damage
// to a creature this turn, the creature dealt damage can't be regenerated
// this turn" (Runesword; CR 701.19c), and ExileIfDealtDamageByWouldDie
// ThisTurnForEffect its sibling "if a creature dealt damage by <dealer>
// would die this turn, exile that creature instead" (CR 614.1a, 700.4).
//
// The dealer is the object the ref names; the effect lasts until cleanup
// whatever becomes of the dealer, because the creatures it damaged are
// what it is about. With fromNow set, only the damage dealt after the
// record is made counts ("if the creature deals damage … this turn");
// otherwise the whole turn's does ("dealt damage by … this turn").
// `controller` controls the replacement; uuid.Nil reads it off the
// source. Registers nothing for a zero ref.
//
// Caller must hold g.mu (write).
func (g *Game) CantBeRegeneratedIfDealtDamageByForEffect(sourceID uuid.UUID, dealer ObjectRef, fromNow bool, label string) bool {
	return g.registerDealtDamageByLocked(sourceID, dealer, uuid.Nil, ModCantBeRegenerated, fromNow, label)
}

// ExileIfDealtDamageByWouldDieThisTurnForEffect: see
// CantBeRegeneratedIfDealtDamageByForEffect.
//
// Caller must hold g.mu (write).
func (g *Game) ExileIfDealtDamageByWouldDieThisTurnForEffect(sourceID uuid.UUID, dealer ObjectRef, controller uuid.UUID, fromNow bool, label string) bool {
	return g.registerDealtDamageByLocked(sourceID, dealer, controller, ModExileIfWouldDie, fromNow, label)
}

func (g *Game) registerDealtDamageByLocked(sourceID uuid.UUID, dealer ObjectRef, controller uuid.UUID, kind ModKind, fromNow bool, label string) bool {
	if dealer.ID == uuid.Nil {
		return false
	}
	m := Mod{Kind: kind, Objects: []ObjectRef{dealer}}
	if fromNow {
		m.DamageSince = g.DamageRecordMarkForEffect()
	}
	return g.appendScopedEffectLocked(sourceID, nil, ScopeDealtDamageBy, controller,
		[]Mod{m}, g.UntilEndOfTurnDuration(), label, timeNowUnixNano())
}

// dealtDamageByModCoversLocked reports whether the ScopeDealtDamageBy
// mod `m` covers the permanent `c`: a creature its dealer has dealt
// damage to this turn, at or after the mod's DamageSince, as the object
// it is now.
func (g *Game) dealtDamageByModCoversLocked(m Mod, c *Card) bool {
	if c == nil || len(m.Objects) != 1 || !c.IsCreature() {
		return false
	}
	return g.objectDealtDamageSinceLocked(m.Objects[0], ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}, m.DamageSince)
}

// dealtDamageByRecordCoversLocked reports whether any `kind` mod of the
// ScopeDealtDamageBy record `e` covers `c`.
func (g *Game) dealtDamageByRecordCoversLocked(e *ScopedEffect, kind ModKind, c *Card) bool {
	if e.Scope != ScopeDealtDamageBy || c == nil {
		return false
	}
	for _, m := range e.Mods {
		if m.Kind == kind && g.dealtDamageByModCoversLocked(m, c) {
			return true
		}
	}
	return false
}

// dealtDamageByScopeProblem is the registration and restore check of the
// dealt-damage-by vocabulary: a ScopeDealtDamageBy record carries only
// cantBeRegenerated and exileIfWouldDie, each naming exactly one dealer
// and a DamageSince that is not negative; and no other record names a
// dealer on those kinds or a DamageSince on any. "" when sound.
func dealtDamageByScopeProblem(scope AffectedScope, mods []Mod) string {
	for _, m := range mods {
		if scope != ScopeDealtDamageBy {
			if m.DamageSince != 0 {
				return fmt.Sprintf("a %s mod with scope %q carries damageSince, which only a dealtDamageBy record reads", m.Kind, scope)
			}
			if (m.Kind == ModCantBeRegenerated || m.Kind == ModExileIfWouldDie) && len(m.Objects) != 0 {
				return fmt.Sprintf("a %s mod with scope %q names objects, which only a dealtDamageBy record reads", m.Kind, scope)
			}
			continue
		}
		if m.Kind != ModCantBeRegenerated && m.Kind != ModExileIfWouldDie {
			return fmt.Sprintf("a dealtDamageBy record carries %s, which it does not read", m.Kind)
		}
		if len(m.Objects) != 1 || m.Objects[0].ID == uuid.Nil {
			return fmt.Sprintf("a dealtDamageBy %s mod needs exactly one dealer", m.Kind)
		}
		if m.DamageSince < 0 {
			return fmt.Sprintf("a dealtDamageBy %s mod has a negative damageSince %d", m.Kind, m.DamageSince)
		}
	}
	return ""
}

// ---------------------------------------------------------------
// Bone Shaman's static: CardDef.DealtDamageCantBeRegenerated
// ---------------------------------------------------------------

// DealtDamageCantBeRegenerated is one "Creatures dealt damage by this
// creature this turn can't be regenerated this turn." static (#1946,
// CR 701.19c) — the ability Bone Shaman gains. Catalog data, never
// stored: the static is read live off the battlefield, so it covers the
// creatures the permanent damaged earlier in the turn too, and it ends
// the moment the permanent leaves or loses the ability.
type DealtDamageCantBeRegenerated struct {
	// Label is the clause as printed.
	Label string
}

// CatalogDealtDamageCantBeRegenerated returns the statics a permanent
// with the given catalog key has. carddef.go sets it from
// CardDef.DealtDamageCantBeRegenerated; a game-package test may stub it.
var CatalogDealtDamageCantBeRegenerated func(key string) []DealtDamageCantBeRegenerated

// dealtDamageCantBeRegeneratedLocked reports whether the permanent `c`
// can't be regenerated because something it was dealt damage by this
// turn says so: a Runesword record naming the dealer, or a battlefield
// permanent with Bone Shaman's static that damaged it.
//
// Caller must hold g.mu.
func (g *Game) dealtDamageCantBeRegeneratedLocked(c *Card) bool {
	if c == nil || len(g.TurnTally.DamagedObjects) == 0 || !c.IsCreature() {
		return false
	}
	for i := range g.ScopedEffects {
		if g.dealtDamageByRecordCoversLocked(&g.ScopedEffects[i], ModCantBeRegenerated, c) {
			return true
		}
	}
	if CatalogDealtDamageCantBeRegenerated == nil || g.Battlefield == nil {
		return false
	}
	victim := ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := catalogAbilityKeyOf(src)
		if key == "" || len(CatalogDealtDamageCantBeRegenerated(key)) == 0 {
			continue
		}
		if g.ObjectDealtDamageToObjectThisTurn(ObjectRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}, victim) {
			return true
		}
	}
	return false
}
