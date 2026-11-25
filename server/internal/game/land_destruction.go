package game

import (
	"slices"

	"github.com/google/uuid"
)

// land_destruction.go — "counter target spell if it would destroy a
// land you control" (Equinox, #6225, ADR 0126's amendment of
// 2026-11-25).
//
// A spell's resolution is catalog code the engine can't read, so
// "would destroy" is answered the only way that is true for every spell
// at once: the spell is resolved on a throwaway copy of the game, and
// the copy is asked what that resolution destroyed. CR 701.8a: "To
// destroy a permanent, move it from the battlefield to its owner's
// graveyard." A destruction is an exit the copy logs with LeaveCause
// "destroy" (leave_cause.go), so an indestructible land (CR 702.12b), a
// regenerated one (CR 701.19a) and a target that is no longer legal
// (CR 608.2b) are not destroyed there either, and a sacrifice (CR
// 701.21a) or a state-based death after the spell (CR 704.5g) is not a
// destruction by it.
//
// The copy is the undo stack's Clone, whose contract is that nothing
// done to it reaches the live game; it is given no listeners, so no
// subscriber hears it, and it is dropped when the answer is read. The
// part of a resolution that waits on a player's choice is not played
// out: the answer is what the spell destroys before its first
// question.

// SpellWouldDestroyALandOfForEffect reports whether the spell on the
// stack with instance ID spellID would destroy a land `player` controls
// if it resolved now. False for an ability and for a spell that has
// left the stack.
//
// Caller must hold g.mu (an effect's resolution).
func (g *Game) SpellWouldDestroyALandOfForEffect(spellID, player uuid.UUID) bool {
	destroyed := g.spellWouldDestroyLocked(spellID)
	for _, id := range destroyed {
		if c := findBattlefieldCard(g, id); c != nil && c.Controller == player && c.IsLand() {
			return true
		}
	}
	return false
}

// spellWouldDestroyLocked is the permanents the spell would destroy if
// it resolved now, played out on a clone with the spell moved to the top
// of the stack. Nil when the spell isn't on the stack.
func (g *Game) spellWouldDestroyLocked(spellID uuid.UUID) (destroyed []uuid.UUID) {
	if item := g.StackMeta[spellID]; item == nil || item.Kind != StackItemSpell {
		return nil
	}
	if _, ok := g.stackCardLocked(spellID); !ok {
		return nil
	}
	sim := g.cloneLocked()
	sim.Listeners = nil
	sim.simulating = true
	// The spell resolves next on the copy: its card to the top of the
	// stack (the last card), and its item stamped after every ability
	// so none resolves first (CR 608.1, resolveTopOfStackLocked).
	idx := slices.IndexFunc(sim.Stack.Cards, func(c Card) bool { return c.InstanceID == spellID })
	if idx < 0 {
		return nil
	}
	card := sim.Stack.Cards[idx]
	sim.Stack.Cards = append(slices.Delete(sim.Stack.Cards, idx, idx+1), card)
	var top uint64
	for _, it := range sim.StackMeta {
		if it != nil && it.Seq > top {
			top = it.Seq
		}
	}
	sim.StackMeta[spellID].Seq = top + 1
	start := len(sim.Events)
	func() {
		// A resolution that fails on the copy destroys what it destroyed
		// before it failed; the live game never runs it this way.
		defer func() { _ = recover() }()
		_ = sim.resolveTopOfStackLocked()
	}()
	for _, ev := range sim.Events[start:] {
		if ev.LeaveCause == LeaveCauseDestroy && ev.CardID != uuid.Nil && !slices.Contains(destroyed, ev.CardID) {
			destroyed = append(destroyed, ev.CardID)
		}
	}
	return destroyed
}
