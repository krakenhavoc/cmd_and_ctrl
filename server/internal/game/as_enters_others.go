package game

import "github.com/google/uuid"

// as_enters_others.go — an "as this enters" clause that changes OTHER
// objects, run as part of the entry itself (#3862, ADR 0082 amendment
// 2026-10-15).
//
//	CR 614.1c  "Effects that read '[This permanent] enters with . . . ,'
//	            'As [this permanent] enters . . . ,' or '[This
//	            permanent] enters as . . . ' are replacement effects."
//
// Ixidron: "As this creature enters, turn all other nontoken creatures
// face down." Its rulings: "You turn the creatures face-down *as*
// Ixidron enters. There is never a moment when Ixidron is on the
// battlefield and the creatures are face-up. If a creature on the
// battlefield has a 'whenever another creature enters' ability, it
// won't trigger because that creature will be face down before Ixidron
// enters." And: "If Ixidron and another creature are entering at the
// same time, the other creature enters face up."
//
// CardDef.AsEnters runs after the entry has been announced, which is
// right for what it is used for (a choice already made, a counter, a
// note on the permanent itself) and wrong for this: the creatures it
// turns over would be face up when the EventETB that announces the
// arrival is harvested, and a Soul Warden among them would trigger.
// So CardDef.AsEntersOthers runs earlier, between the landing and the
// first event about it, and it is told which permanents are arriving in
// the same entry so it can leave them alone (ArrivingForEffect). The
// creatures it changes are changed before anything has seen the new
// permanent, which is what "as it enters" means for them.

// runAsEntersOthersLocked runs the AsEntersOthers clause of every
// permanent an entry landed, before any of them is announced. The
// permanents landing together are "arriving" for the duration, so a
// clause that asks about the battlefield sees it as it was before this
// entry (CR 614.12 looks at what already exists).
//
// A face-down permanent runs nothing (its CatalogKey is silent, CR
// 708.2a), as for its AsEnters.
//
// Caller must hold g.mu.
func (g *Game) runAsEntersOthersLocked(landings []entryLanding) {
	if len(landings) == 0 || CatalogLookup == nil {
		return
	}
	type clause struct {
		id  uuid.UUID
		run func(g *Game, cardID uuid.UUID) error
	}
	var clauses []clause
	for _, l := range landings {
		key := CatalogKey(l.moved)
		if c := findBattlefieldCard(g, l.entered); c != nil && c.TextBox != nil {
			key = catalogKeyOf(c)
		}
		if d := catalogDef(key); d != nil && d.AsEntersOthers != nil {
			clauses = append(clauses, clause{id: l.entered, run: d.AsEntersOthers})
		}
	}
	if len(clauses) == 0 {
		return
	}
	prev := g.arriving
	g.arriving = make([]uuid.UUID, 0, len(landings))
	for _, l := range landings {
		g.arriving = append(g.arriving, l.entered)
	}
	defer func() { g.arriving = prev }()
	for _, c := range clauses {
		if err := c.run(g, c.id); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, Source: c.id, ErrorMsg: err.Error()})
		}
	}
}

// runAsEntersOthersForMoveLocked is runAsEntersOthersLocked for the one
// entry site that lands a card without an entryLanding: the direct move
// (the sandbox drag and the commander's zone replacement).
//
// Caller must hold g.mu.
func (g *Game) runAsEntersOthersForMoveLocked(cardID uuid.UUID) {
	c := findBattlefieldCard(g, cardID)
	if c == nil {
		return
	}
	g.runAsEntersOthersLocked([]entryLanding{{moved: *c, entered: cardID}})
}

// ArrivingForEffect reports whether the permanent `id` is entering the
// battlefield in the entry whose AsEntersOthers clauses are running
// now. Such a permanent is on the battlefield already, but the clause
// must treat it as not there yet: "all other creatures" are the ones
// that were there before (Ixidron's ruling). False outside a clause.
//
// Caller must hold g.mu.
func (g *Game) ArrivingForEffect(id uuid.UUID) bool {
	for _, a := range g.arriving {
		if a == id {
			return true
		}
	}
	return false
}
