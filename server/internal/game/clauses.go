package game

import "github.com/google/uuid"

// clauses.go — #764 / ADR 0065 §§1-2: the ONE walk over an
// announcement's target clauses.
//
// An announcement (a cast at CR 601.2c, a trigger at CR 603.3d, an
// activation at CR 602.2b) has an ordered list of target STEPS. Each
// step is one clause of one mode occurrence, and the player answers
// the steps in order. For the overwhelmingly common case — a
// non-modal card with one clause — the list has exactly one entry
// and everything below is the old behaviour with n == 1.
//
// The three things that read this list:
//
//   - the announce gate (validateAnnouncedTargetsLocked), which
//     checks each pick against ITS OWN clause rather than against a
//     union, which is the whole point of #764;
//   - the CR 608.2b re-check, through clauseForRefLocked, so a
//     second slot that stopped qualifying is skipped at resolution
//     without the card file re-implementing the predicate;
//   - the projection and the pickers (protocol, legal, the client),
//     which walk the same steps to ask the same questions.

// AnnouncedClause is one target step of an announcement: which mode
// occurrence it belongs to, which clause of that occurrence's clause
// list it is, and the clause itself.
type AnnouncedClause struct {
	// Mode is the index into the announced mode list (the
	// OCCURRENCE). 0 for a non-modal announcement.
	Mode int

	// Option is the ModeSpec option index chosen at that occurrence,
	// or -1 for a non-modal announcement. Two occurrences of a
	// repeated mode (CR 700.2d) share an Option and differ in Mode.
	Option int

	// Slot is the clause index within that occurrence's clause list.
	Slot int

	// Clause is the predicate and count for this step, by VALUE with
	// its own Rest cleared: a step is one clause, and holding a copy
	// is what lets the X-count resolution (CountFromX) rewrite a
	// step's Min / Max without touching the catalog's shared spec.
	Clause TargetClause
}

// AnnouncedClauses lists an announcement's target steps in the order
// they are chosen.
//
// `spec` is the card-level / ability-level clause list; when it is
// non-nil it wins outright and the announcement is a single
// occurrence (Mode 0, Option -1). Otherwise the modal path applies:
// one occurrence per entry of `modes`, in announce order, each
// contributing its option's clause list. A modal card with a
// card-level spec (nothing in the catalog declares both, and
// effects.Register refuses it) would use the card-level one.
//
// Pure: takes no lock and touches no game state, so the client's
// step list and the server's gate are demonstrably the same walk.
func AnnouncedClauses(spec *TargetSpec, ms *ModeSpec, modes []int) []AnnouncedClause {
	if spec != nil {
		n := spec.ClauseCount()
		out := make([]AnnouncedClause, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, AnnouncedClause{Mode: 0, Option: -1, Slot: i, Clause: clauseValue(spec, i)})
		}
		return out
	}
	if ms == nil {
		return nil
	}
	var out []AnnouncedClause
	for occ, opt := range modes {
		if opt < 0 || opt >= len(ms.Options) {
			continue
		}
		t := ms.Options[opt].Targets
		for i := 0; i < t.ClauseCount(); i++ {
			out = append(out, AnnouncedClause{Mode: occ, Option: opt, Slot: i, Clause: clauseValue(t, i)})
		}
	}
	return out
}

// itemAnnouncedClauses is the steps an item on the stack was
// announced under. A spell's specs are looked up from the catalog by
// oracle ID (they are catalog data and a restored binary re-derives
// them); an ability's ride the item.
//
// Caller must hold g.mu.
func (g *Game) itemAnnouncedClauses(item *StackItem) []AnnouncedClause {
	if item == nil {
		return nil
	}
	spec, ms := item.targetSpec, item.modeSpec
	if spec == nil && ms == nil && item.Kind == StackItemSpell {
		oracle := g.oracleKeyOfStackItemLocked(item)
		if oracle != "" {
			spec = castTargetSpecForItem(oracle, item)
			ms = ModeSpecFor(oracle)
		}
	}
	return AnnouncedClauses(spec, ms, item.Modes)
}

// oracleKeyOfStackItemLocked is the catalog key of the card behind a
// stack item, or "". Caller must hold g.mu.
func (g *Game) oracleKeyOfStackItemLocked(item *StackItem) string {
	if item == nil || item.SourceCardID == uuid.Nil {
		return ""
	}
	if c := g.findCardByIDLocked(item.SourceCardID); c != nil {
		return CatalogKey(*c)
	}
	return ""
}

// clauseForRefLocked finds the clause a ref was announced under, or
// nil when the item carries no structured clause list (a free-form
// S13.1 announcement, where the re-check is existence only).
//
// Caller must hold g.mu.
func (g *Game) clauseForRefLocked(item *StackItem, ref TargetRef) *TargetClause {
	steps := g.itemAnnouncedClauses(item)
	for i := range steps {
		if steps[i].Mode == ref.Mode && steps[i].Slot == ref.Slot {
			return &steps[i].Clause
		}
	}
	return nil
}

// clauseValue is clause i of a statement as a standalone value: the
// predicate and count, with the statement's tail dropped so nothing
// downstream can mistake it for the clause's own.
func clauseValue(spec *TargetSpec, i int) TargetClause {
	c := *spec.Clause(i)
	c.Rest = nil
	return c
}

// assignAnnouncedSlots stamps (Mode, Slot) onto refs that carry
// none, by filling the steps in printed order, and returns the
// annotated copy.
//
// This is the compatibility half of #764 (ADR 0065 §2). Every ref
// that predates the change — in a snapshot, from an older client,
// from a test fixture, from the bot enumerator — carries the zero
// value for both, and for the overwhelmingly common one-step
// announcement the zero value is already right. The cases it has to
// answer are the two where it is not: a modal card whose TARGETING
// mode is not the first chosen one (Scour for Scrap's "choose one or
// both", where the target belongs to occurrence 1), and a
// multi-clause card whose caller listed the picks in printed order
// without saying so.
//
// A ref that already names a step is left alone, and one ref naming
// a step turns the whole announcement strict — a caller that
// annotates one must annotate all, because a half-annotated list is
// not something the greedy fill can read.
func assignAnnouncedSlots(steps []AnnouncedClause, targets []TargetRef) []TargetRef {
	if len(steps) == 0 || len(targets) == 0 {
		return targets
	}
	for _, t := range targets {
		if t.Mode != 0 || t.Slot != 0 {
			return targets
		}
	}
	out := append([]TargetRef(nil), targets...)
	cursor, filled := 0, 0
	for i := range out {
		if out[i].Kind == TargetSelf || out[i].Kind == TargetNone {
			continue
		}
		// Advance past every step already full. Max 0 is unbounded,
		// so such a step never fills and swallows the rest — which is
		// right: no catalog statement puts an unbounded clause
		// anywhere but last.
		for cursor < len(steps)-1 {
			max := steps[cursor].Clause.Max
			if max > 0 && filled >= max {
				cursor++
				filled = 0
				continue
			}
			break
		}
		out[i].Mode, out[i].Slot = steps[cursor].Mode, steps[cursor].Slot
		filled++
	}
	return out
}

// resolveStepCountsFromX rewrites every X-counted step's Min / Max to
// the announced X (S22's CountFromX, which lives on a clause and so
// may live on a MODE's clause — Heliod's Intervention) and returns
// the indices of the steps it rewrote. The steps hold clause copies,
// so this never touches the catalog.
//
// The returned indices matter because Max 0 means UNBOUNDED
// everywhere else in a clause, so an X of zero cannot be enforced by
// the ordinary count check — announcing X=0 would buy an unbounded
// clause for free, which is the exact shape of the bug CountFromX
// exists to close (#259). The caller checks those steps for an exact
// count of X.
func resolveStepCountsFromX(steps []AnnouncedClause, x int) []int {
	var out []int
	for i := range steps {
		if !steps[i].Clause.CountFromX {
			continue
		}
		steps[i].Clause.Min, steps[i].Clause.Max = x, x
		steps[i].Clause.CountFromX = false
		out = append(out, i)
	}
	return out
}

// stepTargetCount counts the refs that answer one step.
func stepTargetCount(step AnnouncedClause, targets []TargetRef) int {
	n := 0
	for _, t := range targets {
		if t.Kind == TargetSelf || t.Kind == TargetNone {
			continue
		}
		if t.Mode == step.Mode && t.Slot == step.Slot {
			n++
		}
	}
	return n
}

// validateAnnouncedTargetsLocked is the announce-time gate
// (CR 601.2c / 602.2b / 603.3d) over a clause list.
//
// Every ref names a step by (Mode, Slot); a ref naming a step that
// does not exist is ErrInvalidParam. Within a step the count must
// fall within the clause's Min..Max, each pick must be legal under
// THAT clause, and (unless AllowSame) no two picks may name the same
// object. A clause marked Distinct may not repeat any object picked
// for an EARLIER step.
//
// Refs must arrive in step order — the order the picker asks in and
// the order a positional card file reads item.Targets in. Out-of-
// order refs are ErrInvalidParam rather than silently re-sorted: a
// client that sends them is confused about which slot it is filling,
// and quietly fixing it would put the biter in the victim's slot.
//
// Self / None placeholders are skipped, as they always were.
//
// Caller must hold g.mu.
func (g *Game) validateAnnouncedTargetsLocked(caster uuid.UUID, steps []AnnouncedClause, targets []TargetRef) error {
	if len(steps) == 0 {
		// No structured clause list: the S13.1 free-form path, which
		// accepts whatever the client sent.
		return nil
	}
	counts := make([]int, len(steps))
	earlier := make(map[uuid.UUID]bool, len(targets))
	perStep := make([]map[uuid.UUID]bool, len(steps))
	cursor := 0
	for _, t := range targets {
		if t.Kind == TargetSelf || t.Kind == TargetNone {
			continue
		}
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			return ErrInvalidParam
		}
		idx := -1
		for i := cursor; i < len(steps); i++ {
			if steps[i].Mode == t.Mode && steps[i].Slot == t.Slot {
				idx = i
				break
			}
		}
		if idx < 0 {
			return ErrInvalidParam
		}
		// Moving the cursor forward is what enforces step order: a
		// ref for a step already passed can no longer be matched.
		// Every step the cursor leaves behind is finished, so its
		// picks become "earlier" for a later Distinct clause.
		if idx > cursor {
			for i := cursor; i < idx; i++ {
				for id := range perStep[i] {
					earlier[id] = true
				}
			}
			cursor = idx
		}
		clause := &steps[idx].Clause
		if perStep[idx] == nil {
			perStep[idx] = make(map[uuid.UUID]bool, 2)
		}
		if !clause.AllowSame && perStep[idx][t.ID] {
			return ErrInvalidParam
		}
		if clause.Distinct && earlier[t.ID] {
			return ErrInvalidParam
		}
		perStep[idx][t.ID] = true
		counts[idx]++
		if !g.targetLegalLocked(caster, clause, t) {
			return ErrIllegalTarget
		}
	}
	for i := range steps {
		c := &steps[i].Clause
		if counts[i] < c.Min || (c.Max > 0 && counts[i] > c.Max) {
			return ErrInvalidParam
		}
	}
	return nil
}

// anyClauseUnfillableLocked reports whether some REQUIRED clause of
// the list has no legal target at all right now — CR 603.3d's
// "removed from the stack" test for a triggered ability, and the
// gate the cast / activation offer uses.
//
// A clause with Min 0 ("up to one target") is never unfillable: the
// announcement is legal with nothing chosen for it.
//
// Caller must hold g.mu.
func (g *Game) anyClauseUnfillableLocked(caster uuid.UUID, steps []AnnouncedClause) bool {
	for i := range steps {
		c := &steps[i].Clause
		if c.Min <= 0 {
			continue
		}
		lt := g.legalTargetsLocked(caster, c)
		if len(lt.Players)+len(lt.Cards) < c.Min {
			return true
		}
	}
	return false
}
