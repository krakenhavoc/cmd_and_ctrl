package game

import (
	"slices"

	"github.com/google/uuid"
)

// damage_redirected_keyed.go — #4404: a damage event with redirected parts,
// written with its frame.
//
// A charged redirection ("the next 2 damage that would be dealt to you this
// turn is dealt to any target instead", Harm's Way and the Kor family; ADR
// 0108 §9) that meets a bigger event splits it: the charge is dealt to the
// destination and the rest stays where it was. The part rides the event's
// tail (damageTail.redirected) and is dealt, as its own event of the same
// instance, when the event reaches its terminal outcome (CR 614.5: with
// what was already applied to it). So when the rest of the event then waits
// on a CR 616 prompt or a "may" (a Blood of the Martyr offer, a second
// shield), the part waits with it.
//
// Before this the part kept the frame off the disk (damageEventIsData), so
// a table paused there rewound on a deploy. A part is data (its recipient,
// its amount and the effects already applied to it, by ID), so it is now
// written on the tail's mirror (damageTailSnapshot.Redirected). The IDs are
// the same IDs the frame's own marks use, packed over the object serials
// GameSnapshot.ReplacementSerials writes, so they mean the same effects on
// the restored table.
//
// A binary that keys damage but not this would read the frame and drop the
// part: the redirected damage would never be dealt. So an event with parts
// is written under a key of its own, one for noncombat damage and one for
// combat damage, and restore refuses either out of place.

// ReplacementResumeDamageRedirected and ReplacementResumeCombatDamageRedirected
// are the frame keys of a damage event that carries redirected parts
// (#4404), noncombat and combat. On-disk identities: never renamed, never
// reused.
const (
	ReplacementResumeDamageRedirected       = "replacement-resume/damage-redirected"
	ReplacementResumeCombatDamageRedirected = "replacement-resume/combat-damage-redirected"
)

// redirectedDamageSnapshot is a redirectedDamage on disk.
type redirectedDamageSnapshot struct {
	To      uuid.UUID             `json:"to"`
	Amount  int                   `json:"amount"`
	Applied []ReplacementEffectID `json:"applied,omitempty"`
}

func snapshotRedirectedDamage(in []redirectedDamage) []redirectedDamageSnapshot {
	if len(in) == 0 {
		return nil
	}
	out := make([]redirectedDamageSnapshot, 0, len(in))
	for _, p := range in {
		out = append(out, redirectedDamageSnapshot{To: p.to, Amount: p.amount, Applied: slices.Clone(p.applied)})
	}
	return out
}

func restoreRedirectedDamage(in []redirectedDamageSnapshot) []redirectedDamage {
	if len(in) == 0 {
		return nil
	}
	out := make([]redirectedDamage, 0, len(in))
	for _, p := range in {
		out = append(out, redirectedDamage{to: p.To, amount: p.Amount, applied: slices.Clone(p.Applied)})
	}
	return out
}

// damageResumeKey is the key a keyable damage event's frame is written
// under: combat or not, with redirected parts or not.
func damageResumeKey(combat, redirected bool) string {
	switch {
	case combat && redirected:
		return ReplacementResumeCombatDamageRedirected
	case combat:
		return ReplacementResumeCombatDamage
	case redirected:
		return ReplacementResumeDamageRedirected
	}
	return ReplacementResumeDamage
}

// damageEventResumeKey is damageResumeKey for a live event.
func damageEventResumeKey(ev *ReplacementEvent) string {
	return damageResumeKey(damageEventIsCombat(ev), ev.damageTail != nil && len(ev.damageTail.redirected) > 0)
}

// snapshotDamageResumeKey is damageResumeKey for a written event: the one
// key restore accepts it under.
func snapshotDamageResumeKey(ev *replacementEventSnapshot) string {
	return damageResumeKey(snapshotDamageIsCombat(ev), ev.DamageTail != nil && len(ev.DamageTail.Redirected) > 0)
}
