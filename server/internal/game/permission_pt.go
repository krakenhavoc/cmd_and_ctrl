package game

// permission_pt.go — "cast creature spells with power or toughness N or
// less" (Leonardo, Sewer Samurai, #4964; ADR 0066 amendment 2026-11-04).
//
// PermissionFilter is pure data and Matches is handed a Card, not the
// game. Every clause it had read something the card carries. Power and
// toughness are different for one kind of card: a `*` whose number a
// characteristic-defining ability supplies. That ability functions in
// all zones (CR 604.3, CR 113.6a, CR 208.2a), so Tarmogoyf in a graveyard
// is a creature card with the power its ability counts there, and the
// count needs the game (the graveyards, the hand, the battlefield).
//
// So the lookups that hold the game LEND the answer to the copy they
// judge, as ThroughTextOf lends a text to a target predicate: one
// reading per face (Card.ptAnywhere), worked out by cdaAnywhereLocked,
// the reader Talara's Bane and craft already use. The filter then reads
// the reading for the face it is judging (a modal card's two creature
// faces are two answers, CR 712.11c). Nothing is written to a card in a
// zone, so nothing goes stale and nothing is persisted.
//
// A card with such an ability that reaches the filter WITHOUT a reading
// (a caller that did not lend one), and a `*` the engine cannot compute
// at all (no catalog ability, Card.VariableToughness), do not qualify:
// a number the engine does not know is never "N or less", so a missed
// lookup makes the permission narrower than printed, never wider.
//
// The card is judged where it is, as the cast is proposed (CR 601.2,
// CR 601.3): a graveyard card's ability counts with the card still in
// the graveyard.

// ptReading is one face's power and toughness off the battlefield, with
// its characteristic-defining abilities applied. known is false for a
// face that is not a creature (CR 208.3) or whose `*` the engine cannot
// compute.
type ptReading struct {
	power, toughness int
	known            bool
}

// hasPTDefiningAbility reports whether the card, as the face it shows,
// has a catalogued layer-7a ability (CR 604.3, CR 613.4a): a `*` the
// engine can compute.
func hasPTDefiningAbility(c *Card) bool {
	if CatalogStaticAbilities == nil {
		return false
	}
	for _, s := range CatalogStaticAbilities(catalogAbilityKeyOf(c)) {
		if s.Layer == Layer7PT && s.SubLayer == SubLayer7A_CDA && s.Apply != nil {
			return true
		}
	}
	return false
}

// facesWithPTDefiningAbility reports whether any face of the card has a
// layer-7a ability.
func facesWithPTDefiningAbility(c Card) bool {
	if len(c.Faces) < 2 {
		return hasPTDefiningAbility(&c)
	}
	for i := range c.Faces {
		f := c
		f.SetFace(i)
		if hasPTDefiningAbility(&f) {
			return true
		}
	}
	return false
}

// withPTAnywhereLocked lends c a reading of its power and toughness per
// face (see the file comment). It returns c unchanged for a permanent
// (the layer pass has applied its abilities already), for a copy that
// already has one, and for a card with no layer-7a ability on any face,
// whose printed numbers are the answer. c is a value; the card in its
// zone is not written.
//
// Caller must hold g.mu (read or write).
func (g *Game) withPTAnywhereLocked(c Card) Card {
	if c.effective != nil || c.ptAnywhere != nil || !facesWithPTDefiningAbility(c) {
		return c
	}
	n := c.FaceCount()
	readings := make([]ptReading, n)
	for i := 0; i < n; i++ {
		f := c
		f.SetFace(i)
		readings[i] = g.ptReadingLocked(f)
	}
	c.ptAnywhere = readings
	return c
}

// ptReadingLocked is one face's reading: f has that face materialised.
//
// Caller must hold g.mu.
func (g *Game) ptReadingLocked(f Card) ptReading {
	if !f.IsCreature() {
		return ptReading{}
	}
	if hasPTDefiningAbility(&f) {
		p, t := ptWithCounters(g.cdaAnywhereLocked(f), f.Counters)
		return ptReading{power: p, toughness: t, known: true}
	}
	if f.VariableToughness {
		return ptReading{}
	}
	p, t := ptWithCounters(effectiveOf(&f), f.Counters)
	return ptReading{power: p, toughness: t, known: true}
}

// powerOrToughnessAtMost is the filter's clause: c is a creature whose
// power or toughness is n or less, judged on the face c shows.
func powerOrToughnessAtMost(c *Card, n int) bool {
	r := ptForFilter(c)
	return r.known && (r.power <= n || r.toughness <= n)
}

// powerAtLeast is the filter's clause: c is a creature whose power is n or
// more, judged on the face c shows.
func powerAtLeast(c *Card, n int) bool {
	r := ptForFilter(c)
	return r.known && r.power >= n
}

// ptForFilter is the card's power and toughness for a filter: the layer
// pass's numbers on the battlefield, the lent reading for the face it
// shows when it has one, and otherwise its printed numbers, unknown for
// a `*` (see the file comment).
func ptForFilter(c *Card) ptReading {
	if !c.IsCreature() {
		return ptReading{}
	}
	if c.effective != nil {
		p, t := ptWithCounters(effectiveOf(c), c.Counters)
		return ptReading{power: p, toughness: t, known: true}
	}
	if c.ptAnywhere != nil {
		if c.ActiveFace < 0 || c.ActiveFace >= len(c.ptAnywhere) {
			return ptReading{}
		}
		return c.ptAnywhere[c.ActiveFace]
	}
	if c.VariableToughness || hasPTDefiningAbility(c) {
		return ptReading{}
	}
	p, t := ptWithCounters(effectiveOf(c), c.Counters)
	return ptReading{power: p, toughness: t, known: true}
}

// lendPTLocked is withPTAnywhereLocked for a lookup asking permission p
// about c: only a permission with the power-or-toughness clause needs a
// reading, so every other permission costs nothing more.
//
// Caller must hold g.mu.
func (g *Game) lendPTLocked(p *CastPermission, c Card) Card {
	if p == nil || !(p.Filter.PowerOrToughnessCapped || p.Filter.PowerAtLeastCapped) {
		return c
	}
	return g.withPTAnywhereLocked(c)
}

// permissionForSpellLocked is ForSpell with the spell's power and
// toughness lent first, for the cast paths that judge a permission
// against the face being cast.
//
// Caller must hold g.mu.
func (g *Game) permissionForSpellLocked(p *CastPermission, spell Card) *CastPermission {
	return p.ForSpell(g.lendPTLocked(p, spell))
}
