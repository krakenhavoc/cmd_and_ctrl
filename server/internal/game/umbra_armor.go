package game

import "github.com/google/uuid"

// umbra_armor.go — CR 702.89, umbra armor (printed "totem armor" before
// the rename, CR 702.89b).
//
// CR 702.89a: "Umbra armor is a static ability that appears on some
// Auras. 'Umbra armor' means 'If enchanted permanent would be destroyed,
// instead remove all damage marked on it and destroy this Aura.'"
//
// It is a replacement effect printed on an Aura, so it is a catalog
// replacement declared on the Aura's Spec (effects.UmbraArmor), not an
// engine built-in: it has a source, and it stops when the Aura leaves or
// loses its abilities. These are the engine halves it calls.
//
// What it replaces, and what it does not:
//
//   - A DESTRUCTION of the enchanted permanent, read off the event the
//     same way regeneration reads it (ev.Destruction): a destroy spell,
//     a wipe, lethal damage (CR 704.5g) and deathtouch damage (CR
//     704.5h). Sacrifice, toughness 0 or less (CR 704.5f), the legend
//     rule, exile and bounce are not destructions and are not replaced.
//   - It is NOT a regeneration, so "can't be regenerated" (CR 701.19c)
//     does not stop it: the gate here never asks
//     regenerationRefusedLocked.
//   - It does not tap the permanent or remove it from combat; it
//     removes the damage (through clearBattlefieldDamage, so the CR
//     704.5g/h check does not fire again on the next pass) and destroys
//     the Aura. An indestructible Aura survives its own "destroy this
//     Aura"; the permanent is still saved, because the replacement has
//     already happened.
//
// Ordering. The affected permanent's controller orders it against a
// regeneration shield or another umbra (CR 616.1). Two copies of one
// umbra destroy two different Auras, so the effect declares
// ActsOnItsSource and two Bear Umbras prompt instead of collapsing.

// UmbraArmorAppliesForEffect is umbra armor's AppliesTo: ev is a
// destruction of the permanent the Aura `auraID` is attached to, and
// the Aura is on the battlefield.
//
// Caller must hold g.mu.
func (g *Game) UmbraArmorAppliesForEffect(ev *ReplacementEvent, auraID uuid.UUID) bool {
	if ev == nil || ev.Kind != RepEventMove || !ev.Destruction {
		return false
	}
	aura := findBattlefieldCard(g, auraID)
	if aura == nil || aura.AttachedTo.Kind != TargetCard || aura.AttachedTo.ID != ev.CardID {
		return false
	}
	return findBattlefieldCard(g, ev.CardID) != nil
}

// UmbraArmorReplaceForEffect is umbra armor's Replace (CR 702.89a): the
// destruction is cancelled, all damage marked on the enchanted
// permanent is removed, and the Aura `auraID` is destroyed.
//
// Caller must hold g.mu (write).
func (g *Game) UmbraArmorReplaceForEffect(ev *ReplacementEvent, auraID uuid.UUID) error {
	if ev == nil {
		return nil
	}
	c := findBattlefieldCard(g, ev.CardID)
	if c == nil {
		return nil
	}
	ev.Cancel()
	clearBattlefieldDamage(c)
	return g.destroyBattlefieldPermanentLocked(auraID, DestroyOptions{})
}

// --- umbra armor granted to Auras (#5457) -------------------------------
//
// "Auras attached to permanents you control have umbra armor" (Umbra
// Mystic). Each such Aura has its own umbra armor, so a permanent with
// two Auras is saved by one of them and its controller chooses which is
// destroyed (CR 616.1). One replacement on the granting permanent could
// not say that, so the grant is declared as data
// (ReplacementEffect.GrantsUmbraArmor) and the gather makes one umbra
// armor replacement per Aura attached to the permanent that would be
// destroyed.
//
//   - An Aura that already has umbra armor of its own gets no second one:
//     both would destroy the same Aura.
//   - Two grants (two Umbra Mystics) give an Aura one replacement, for the
//     same reason.
//   - A granting permanent destroyed in the same event still grants: its
//     ability existed when the event began (CR 614.4), read off the
//     batch's pre-move copy as every departed source's is (#4521,
//     simultaneous_replacements.go). So a Wrath of God that destroys the
//     Mystic and an enchanted creature together destroys the Aura instead
//     of the creature.
//
// IDs. One per Aura, by its object serial, in a range of its own above
// the graveyard replacements' and inside the self-replacement range's
// headroom. The gather is deterministic, so a paused frame regathers it
// by ID like any other.

// UmbraArmorLabel is the name umbra armor carries into a CR 616 prompt.
const UmbraArmorLabel = "Umbra armor"

const (
	grantedUmbraArmorIDBase = graveyardReplacementIDBase + graveyardReplacementIDSpan
	grantedUmbraArmorIDSpan = 1 << 39
)

// umbraArmorGrantedLocked reports whether some permanent's
// GrantsUmbraArmor effect covers the permanent ev would destroy: a
// permanent on the battlefield, or a member of ev's simultaneous exit
// that has already left.
//
// Caller must hold g.mu.
func (g *Game) umbraArmorGrantedLocked(ev *ReplacementEvent) bool {
	if CatalogReplacements == nil || ev == nil || ev.Kind != RepEventMove || !ev.Destruction {
		return false
	}
	grants := func(card *Card) bool {
		reps := CatalogReplacements(catalogAbilityKeyOf(card))
		for i := range reps {
			if !reps[i].GrantsUmbraArmor || !reps[i].ActiveWhen.Active(*card) {
				continue
			}
			if reps[i].AppliesTo == nil || reps[i].AppliesTo(ev, g, card) {
				return true
			}
		}
		return false
	}
	for i := range g.Battlefield.Cards {
		if grants(&g.Battlefield.Cards[i]) {
			return true
		}
	}
	for _, c := range g.departedBatchSourcesLocked(ev) {
		card := c
		if grants(&card) {
			return true
		}
	}
	return false
}

// hasOwnUmbraArmorLocked reports whether the Aura's own abilities include
// umbra armor.
func hasOwnUmbraArmorLocked(aura *Card) bool {
	if CatalogReplacements == nil {
		return false
	}
	reps := CatalogReplacements(catalogAbilityKeyOf(aura))
	for i := range reps {
		if reps[i].UmbraArmor && reps[i].ActiveWhen.Active(*aura) {
			return true
		}
	}
	return false
}

// gatherGrantedUmbraArmorLocked appends one umbra armor replacement per
// Aura attached to the permanent ev would destroy, when a grant covers
// that permanent and the Aura has none of its own.
//
// Caller must hold g.mu.
func (g *Game) gatherGrantedUmbraArmorLocked(ev *ReplacementEvent, applied map[ReplacementEffectID]bool, serials *replacementSerialMinter, out []activeReplacement) []activeReplacement {
	if !g.umbraArmorGrantedLocked(ev) {
		return out
	}
	for i := range g.Battlefield.Cards {
		aura := &g.Battlefield.Cards[i]
		if aura.AttachedTo.Kind != TargetCard || aura.AttachedTo.ID != ev.CardID || !aura.HasSubtype("Aura") {
			continue
		}
		if hasOwnUmbraArmorLocked(aura) {
			continue
		}
		raw := ReplacementEffectID(serials.serial(aura))
		if raw >= grantedUmbraArmorIDSpan {
			continue
		}
		id := grantedUmbraArmorIDBase + raw
		if applied[id] {
			continue
		}
		out = append(out, activeReplacement{
			effect:    grantedUmbraArmorReplacement(aura.InstanceID),
			source:    aura,
			sourceRef: ObjectRef{ID: aura.InstanceID, Epoch: aura.ObjectEpoch},
			id:        id,
		})
	}
	return out
}

// grantedUmbraArmorReplacement is the umbra armor of one Aura that was
// granted it. It applies while the grant does.
func grantedUmbraArmorReplacement(auraID uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
			return g.UmbraArmorAppliesForEffect(ev, auraID) && g.umbraArmorGrantedLocked(ev)
		},
		Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
			return umbraArmorReplace(g, ev, auraID)
		},
		Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
			return g.AffectedPermanentControllerForEffect(ev)
		},
		ActsOnItsSource: true,
		UmbraArmor:      true,
		Label:           UmbraArmorLabel,
	}
}

// umbraArmorReplace is UmbraArmorReplaceForEffect, reached from the
// gather. Assigned in init: a direct call would be an initialisation
// cycle through the exit primitives, which every keyed tail's initialiser
// reaches.
var umbraArmorReplace func(g *Game, ev *ReplacementEvent, auraID uuid.UUID) error

func init() {
	umbraArmorReplace = func(g *Game, ev *ReplacementEvent, auraID uuid.UUID) error {
		return g.UmbraArmorReplaceForEffect(ev, auraID)
	}
}

// grantedUmbraArmorMetaLocked names a granted umbra armor for the prompt
// view: the label and the Aura, found by the serial the ID carries. An
// Aura that has left names nothing.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) grantedUmbraArmorMetaLocked(id ReplacementEffectID) (label string, card uuid.UUID, ok bool) {
	if id < grantedUmbraArmorIDBase || id >= grantedUmbraArmorIDBase+grantedUmbraArmorIDSpan {
		return "", uuid.Nil, false
	}
	serial := int(id - grantedUmbraArmorIDBase)
	for i := range g.Battlefield.Cards {
		if s, known := g.replacementObjectSerialOfLocked(&g.Battlefield.Cards[i]); known && s == serial {
			return UmbraArmorLabel, g.Battlefield.Cards[i].InstanceID, true
		}
	}
	return UmbraArmorLabel, uuid.Nil, true
}

// AffectedPermanentControllerForEffect is the CR 616.1 chooser for a
// replacement of something happening to a permanent: that permanent's
// controller. uuid.Nil when it is not on the battlefield.
//
// Caller must hold g.mu.
func (g *Game) AffectedPermanentControllerForEffect(ev *ReplacementEvent) uuid.UUID {
	if ev == nil {
		return uuid.Nil
	}
	return g.controllerOfBattlefieldCardLocked(ev.CardID)
}

// hostsBeforeAttachmentsLocked reorders a set of permanents about to be
// destroyed together so that every permanent comes before the Auras and
// Equipment in the set that are attached to it, keeping the order
// otherwise.
//
// The batch's moves are sequential (simultaneous.go), but the
// destruction is one event: an umbra destroyed in the same wipe as its
// creature is still on the battlefield when the creature's destruction
// is replaced (CR 702.89a, CR 616.1), so the creature survives and the
// Aura is destroyed. Moving the Aura's leg first would take the
// replacement away; moving it after its host lets the host's leg apply
// it, and the Aura's own leg then finds it gone.
//
// Caller must hold g.mu.
func (g *Game) hostsBeforeAttachmentsLocked(ids []uuid.UUID) []uuid.UUID {
	if len(ids) < 2 {
		return ids
	}
	in := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		in[id] = true
	}
	hosts := make([]uuid.UUID, 0, len(ids))
	var attached []uuid.UUID
	for _, id := range ids {
		if c := findBattlefieldCard(g, id); c != nil && c.AttachedTo.Kind == TargetCard && in[c.AttachedTo.ID] {
			attached = append(attached, id)
			continue
		}
		hosts = append(hosts, id)
	}
	if len(attached) == 0 {
		return ids
	}
	return append(hosts, attached...)
}
