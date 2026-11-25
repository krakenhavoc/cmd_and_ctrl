package game

import "github.com/google/uuid"

// damaged_by_source.go — S82 OLD batch 254 (#5536): a PLAYER target judged
// against the source object, "target player who was dealt combat damage by
// this creature this turn" (Hope of Ghirapur).
//
// PlayerOK is handed the caster and the candidate, never the object whose
// ability this is, so the clause carries it as DATA
// (TargetSpec.DealtCombatDamageBySource), judged beside PlayerOK in the one
// player walk, so the legal set, the announce check (CR 601.2c) and the
// resolution re-check (CR 608.2b) are one answer. The source is
// TargetSource.ObjectRef: the live object while it stands and its last-known
// record once it has left (CR 608.2h), which is how a sacrifice paid as the
// cost is still the Thopter that dealt the damage. With no source object the
// clause admits nobody, never everybody.

// playerDamagedBySourceAdmits is DealtCombatDamageBySource's gate on one
// player candidate: true without the flag; with it, a player the source
// object dealt combat damage to this turn.
//
// Caller must hold g.mu.
func (s *TargetSpec) playerDamagedBySourceAdmits(g *Game, src TargetSource, p *Player) bool {
	if !s.DealtCombatDamageBySource {
		return true
	}
	ref, ok := src.ObjectRef()
	if !ok || ref.ID == uuid.Nil || p == nil {
		return false
	}
	return g.ObjectRefDealtCombatDamageToPlayerThisTurn(ObjectRef(ref), p.ID)
}
