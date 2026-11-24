package game

import (
	"fmt"

	"github.com/google/uuid"
)

// only_these_can_attack.go — "Only the chosen creatures can attack during
// that combat phase" (Last Night Together, #6087).
//
// A live-scope restriction (CR 611.2c: "can't attack" changes no
// characteristic, so the set is read as the rule is asked): "can't
// attack" on every creature EXCEPT the objects the record names. A
// creature that arrives later, whoever controls it, is covered, and so is
// a chosen creature that left the battlefield and came back, because it
// is a new object (CR 400.7) and the record names the object by its
// epoch. Pinning "can't attack" to the creatures on the battlefield as
// the spell resolves would miss the late arrivals.
//
// The objects ride the restriction mod's Objects, as ScopeDealtDamageBy's
// dealer does, so the record stays plain data. It is bound to one combat
// phase through DuringCombatPhaseRuleEffectForEffect (ADR 0188 Decision
// 13), and lasts until the end of that combat.

// ScopeCreaturesExceptObjects is "every creature except the objects its
// mods name", read live (#6087). It carries only addRestrictions mods,
// each naming at least one object (exceptObjectsScopeProblem); the
// excepted set is the union of their Objects. An on-disk identity: never
// renamed, never reused.
const ScopeCreaturesExceptObjects AffectedScope = "creaturesExceptObjects"

// OnlyTheseCanAttackMod is "only <objects> can attack": "can't attack"
// for a ScopeCreaturesExceptObjects record, sparing `objects`.
func OnlyTheseCanAttackMod(objects []ObjectRef) Mod {
	return Mod{Kind: ModAddRestrictions, Restrictions: CantAttack, Objects: cloneObjectRefs(objects)}
}

// exceptObjectsPredicate is the AppliesTo of a ScopeCreaturesExceptObjects
// record: every creature that is not one of the objects its mods name.
func exceptObjectsPredicate(mods []Mod) func(*Card, *Game, *Card) bool {
	var except []ObjectRef
	for _, m := range mods {
		except = append(except, m.Objects...)
	}
	return func(target *Card, _ *Game, _ *Card) bool {
		if target == nil || !target.IsCreature() {
			return false
		}
		for _, o := range except {
			if target.InstanceID == o.ID && target.ObjectEpoch == o.Epoch {
				return false
			}
		}
		return true
	}
}

// exceptObjectsScopeProblem is the registration and restore check for
// ScopeCreaturesExceptObjects: only addRestrictions mods, each naming at
// least one object; and no addRestrictions mod names objects under any
// other scope. "" when sound.
func exceptObjectsScopeProblem(scope AffectedScope, mods []Mod) string {
	for _, m := range mods {
		if scope != ScopeCreaturesExceptObjects {
			if m.Kind == ModAddRestrictions && len(m.Objects) != 0 {
				return fmt.Sprintf("an addRestrictions mod with scope %q names objects, which only a creaturesExceptObjects record reads", scope)
			}
			continue
		}
		if m.Kind != ModAddRestrictions {
			return fmt.Sprintf("a creaturesExceptObjects record carries %s, which it does not read", m.Kind)
		}
		if len(m.Objects) == 0 {
			return "a creaturesExceptObjects mod names no object"
		}
		for _, o := range m.Objects {
			if o.ID == uuid.Nil {
				return "a creaturesExceptObjects mod names a nil object"
			}
		}
	}
	return ""
}

// OnlyTheseCanAttackDuringCombatPhaseForEffect is "Only the chosen
// creatures can attack during that combat phase" (Last Night Together):
// a ScopeCreaturesExceptObjects "can't attack" bound to combat phase
// `phaseID` of this turn, sparing each of `ids` that is on the
// battlefield now, as the object it is now (CR 400.7). `controller` is
// the effect's "you".
//
// Returns false, writing nothing, when none of `ids` is on the
// battlefield or the phase is not a combat phase still ahead
// (DuringCombatPhaseRuleEffectForEffect).
//
// Caller must hold g.mu (write).
func (g *Game) OnlyTheseCanAttackDuringCombatPhaseForEffect(phaseID int, sourceID, controller uuid.UUID, ids []uuid.UUID, label string) bool {
	var refs []ObjectRef
	for _, id := range ids {
		if c := findBattlefieldCard(g, id); c != nil {
			refs = append(refs, ObjectRef{ID: id, Epoch: c.ObjectEpoch})
		}
	}
	if len(refs) == 0 {
		return false
	}
	return g.DuringCombatPhaseRuleEffectForEffect(phaseID, sourceID, ScopeCreaturesExceptObjects, controller,
		[]Mod{OnlyTheseCanAttackMod(refs)}, label)
}
