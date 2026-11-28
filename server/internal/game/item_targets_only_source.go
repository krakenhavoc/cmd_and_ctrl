package game

import "github.com/google/uuid"

// item_targets_only_source.go — Silver Wyvern's "target spell or ability that
// targets only this creature": a stack clause judged against the clause's own
// source (TargetSpec.ItemTargetsOnlySource). The stack item named by itemID
// qualifies when it has at least one target and every target is the source
// object. A spell's item id is its card's instance id.

// itemTargetsOnlySourceAdmits is ItemTargetsOnlySource's (and, looser,
// ItemTargetsSource's) gate on one stack item: true without either flag; with
// ItemTargetsOnlySource, true only for an item whose targets are all the
// walk's source; with ItemTargetsSource, true for an item that has the source
// among its targets. A walk with no source, and an id that is not on the
// stack, admit nothing.
//
// Caller must hold g.mu.
func (s *TargetSpec) itemTargetsOnlySourceAdmits(g *Game, src TargetSource, itemID uuid.UUID) bool {
	if s == nil || (!s.ItemTargetsOnlySource && !s.ItemTargetsSource) {
		return true
	}
	self := src.DepartedID
	if src.Object != nil {
		self = src.Object.InstanceID
	}
	item := g.StackMeta[itemID]
	if self == uuid.Nil || item == nil {
		return false
	}
	if s.ItemTargetsSource {
		// "targets this creature": the source is among the targets (CR 115.1).
		for _, t := range item.Targets {
			if t.Kind == TargetCard && t.ID == self {
				return true
			}
		}
		return false
	}
	n := 0
	for _, t := range item.Targets {
		if t.Kind != TargetCard && t.Kind != TargetPlayer {
			continue
		}
		if t.Kind != TargetCard || t.ID != self {
			return false
		}
		n++
	}
	return n > 0
}
