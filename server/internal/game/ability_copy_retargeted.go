package game

import "github.com/google/uuid"

// ability_copy_retargeted.go — "copy that ability for each other creature
// you control that ability could target. Each copy targets a different one
// of those creatures" (Agrus Kos, Eternal Soldier, #6094): the ability twin
// of spell_copy_retargeted.go (Precursor Golem, Mirrorwing Dragon, Zada).
//
// Nobody chooses the copies' targets: every target of the original that is
// the object `from` is the object `to` in the copy, and every other target
// is kept (CR 707.10: a copy copies the choices made for the original). A
// copy is made only for a `to` the ability could target: the whole new list
// must pass the announce gate, judged for the copy's controller and from the
// ability's source as a copy's targets are judged (copyTargetSourceLocked,
// ADR 0072 §2), so a creature with hexproof from that player or protection
// from the source's colour is skipped.
//
// The ability must still be on the stack: the clause says "that ability",
// and an ability countered in response is gone (CopyAbilityForEffect's
// posture). One that can't be copied (Gogo) makes no copy.

// retargetedAbilityCopyLocked builds the copy's target list, reporting false
// when there is no ability to copy, it can't be copied, nothing of it is
// `from`, or the new list is not legal. Caller must hold g.mu.
func (g *Game) retargetedAbilityCopyLocked(itemID, controller, from, to uuid.UUID) (*StackItem, []TargetRef, bool) {
	item, ok := g.stackAbilityLocked(itemID)
	if !ok || item.Kind == StackItemSpell || item.Uncopyable {
		return nil, nil, false
	}
	next := make([]TargetRef, len(item.Targets))
	moved := false
	for i, t := range item.Targets {
		if (t.Kind == TargetCard || t.Kind == TargetPlayer) && t.ID == from {
			// #6159, CR 707.10d: `to` may be a player where `from` is a
			// permanent, or the other way round (Radiant Performer).
			t.Kind, t.ID = g.retargetKindLocked(to), to
			moved = true
		}
		next[i] = t
	}
	if !moved {
		return nil, nil, false
	}
	cf := &copyFrame{src: g.abilitySourceCardLocked(item), item: *item, controller: controller}
	if err := g.validateAnnouncedTargetsWithLocked(g.copyTargetSourceLocked(cf), g.itemAnnouncedClauses(item), next, nil, item.Targets); err != nil {
		return nil, nil, false
	}
	return item, next, true
}

// AbilityCouldTargetInsteadForEffect reports whether a copy of the ability
// `itemID`, controlled by `controller`, could have every target that is
// `from` be `to` instead: Agrus Kos's "each other creature you control that
// ability could target". Caller must hold g.mu.
func (g *Game) AbilityCouldTargetInsteadForEffect(itemID, controller, from, to uuid.UUID) bool {
	_, _, ok := g.retargetedAbilityCopyLocked(itemID, controller, from, to)
	return ok
}

// CopyAbilityRetargetedEachForEffect puts a copy of the ability `itemID` on
// the stack under `controller` for each `to` it could target instead of
// `from`, in the order given, each copy's every target that is `from` being
// that `to`. Nothing is asked. It reports how many copies it made. Like
// every ability copy, a copy is created, not activated or triggered (CR
// 707.10a), and the objects it targets become targets (CR 115.3).
//
// Caller must hold g.mu.
func (g *Game) CopyAbilityRetargetedEachForEffect(itemID, controller, from uuid.UUID, tos []uuid.UUID) int {
	made := 0
	for _, to := range tos {
		item, next, ok := g.retargetedAbilityCopyLocked(itemID, controller, from, to)
		if !ok {
			continue
		}
		g.createAbilityCopyLocked(item, controller, next)
		made++
	}
	return made
}
