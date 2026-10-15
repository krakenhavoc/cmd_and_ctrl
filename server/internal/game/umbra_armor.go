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
