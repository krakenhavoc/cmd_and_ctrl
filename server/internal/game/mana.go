package game

import "github.com/google/uuid"

// mana.go is the per-player mana economy: a `ManaPool` slice of
// individual `ManaToken` entries (one per produced unit), with the
// helpers the cost validator + auto-tapper consume. The pool is a
// SLICE, not a multiset, so:
//   - Insertion order is observable (S15+ pain-source preview, S17+
//     filter-land sub-payment routing).
//   - Each token can carry its own restrictions independently of
//     other same-color tokens (Cavern of Souls' tribe restriction
//     when it lands, snow-source tagging if S17 enforces it).
//
// The pool is per-player. Pool-wide invariants (CR 106.4 "empty at
// end of step") are enforced from the step-advance path, not on the
// pool struct itself.
//
// Added in S15 sub-PR 2.

// ManaToken is one unit of mana sitting in a player's pool. Color
// is one of "W"/"U"/"B"/"R"/"G"/"C". Source is the permanent that
// produced it (uuid.Nil when produced by a non-card affordance like
// the admin "give mana" debug action that may land later).
// Restrictions is a list of opaque tags the cost validator can
// honour ("cast_creature_only", "spend_on_X_type"). Empty for
// vanilla basics + Sol Ring; populated by S22's Cavern-of-Souls-
// style cards.
type ManaToken struct {
	Color        string
	Source       uuid.UUID
	Restrictions []string

	// SourceKinds is what the producing permanent WAS when this mana
	// was made (#1212, mana_source.go): snow, Treasure, creature,
	// land, artifact, enchantment.
	//
	// A snapshot rather than a second look through Source, because
	// the commonest reader asks about a source that is gone. A
	// Treasure sacrifices itself to pay for its own mana ability, so
	// "if mana from a Treasure was spent to cast it" is asked of a
	// permanent that has ceased to exist (CR 111.7) — the uuid names
	// nothing and a lookup would answer no. Zero for a mana that came
	// from something that is not a permanent.
	SourceKinds ManaSourceKinds

	// Riders are what this mana does WHEN IT IS SPENT (#1547,
	// mana_spend_rider.go): Cavern of Souls' "that spell can't be
	// countered", Pyromancer's Goggles' copy, Hall of the Bandit Lord's
	// haste. Copied from the ability at mint, fired when the token pays
	// for a spell or ability its filter admits, and stamped Applied on
	// the copy in StackItem.Paid.Mana. Nil for ordinary mana.
	Riders []ManaSpendRider `json:"riders,omitempty"`
}

// ManaPool is a player's current mana pool. Order matters — see the
// file-level comment. nil and len-0 are equivalent; helpers preserve
// nil-when-empty so JSON encoding stays sparse.
type ManaPool []ManaToken

// AddMana appends one or more tokens to the pool. Empty tokens
// (zero Color) are dropped on the floor — defensive.
func (p *ManaPool) AddMana(tokens ...ManaToken) {
	for _, t := range tokens {
		if t.Color == "" {
			continue
		}
		*p = append(*p, t)
	}
}

// EmptyPool drops every token. Used by the step-advance hook
// (CR 106.4: pools empty at end of step / phase). Returns the count
// of tokens cleared so callers can emit one event per drop if they
// want (the engine emits a single EventManaPoolEmptied per affected
// player today).
func (p *ManaPool) EmptyPool() int {
	n := len(*p)
	*p = nil
	return n
}

// CanPay reports whether the pool currently holds enough mana to
// cover `cost`, spending only UNRESTRICTED tokens. Dry-run; does not
// mutate the pool.
//
// Prefer CanPayFor: this shorthand passes the zero spend context,
// which no restriction matches, so a pool of Ancient Ziggurat mana
// reads as empty here. That is the safe default for a caller with no
// object in hand — see mana_restriction.go — but every real payment
// path knows what it is paying for and should say so.
//
// `xValue` is the announce-time X value the caller picked; the
// pool needs cost.Generic + cost.XSlots*xValue generic-tradeable
// mana. Pass 0 for spells without X.
func (p ManaPool) CanPay(cost ParsedCost, xValue int) bool {
	return p.CanPayFor(cost, xValue, ManaSpendContext{})
}

// CanPayFor is CanPay with a spend context — the object the mana is
// being paid for. Tokens whose Restrictions the context does not
// satisfy are invisible to the solver.
//
// Algorithm: greedy, restriction-first. For each ColorRequirement in
// cost.Required, find the most-restricted SPENDABLE matching token
// (the token with the fewest remaining ways to be useful), so a
// creature spell paid out of a pool holding both a Ziggurat {G} and a
// Forest {G} burns the Ziggurat mana — which is the play a human
// makes, and the only one that doesn't waste it. Then satisfy generic
// from whatever's left, taking colorless first to preserve colored
// for later costs.
//
// S32 / #352: before this, the restriction-first wording in the
// comment described an intent the code never had — every token
// compared equal because nothing ever set Restrictions.
func (p ManaPool) CanPayFor(cost ParsedCost, xValue int, ctx ManaSpendContext) bool {
	_, _, ok := p.attemptSpend(cost, xValue, ctx, SpendPreserveColors)
	return ok
}

// ManaSpendStrategy chooses HOW the solver pays the generic half of a
// cost when it has a choice (#761). It never changes WHETHER a cost is
// payable: the coloured-requirement pass is identical under both
// strategies, and the generic pass walks every colour bucket either
// way — only the order within it differs. A solver whose preferences
// could make a payable cost unpayable is one nobody can reason about.
type ManaSpendStrategy uint8

const (
	// SpendPreserveColors is the default and the behaviour every
	// payment had before #761: colourless first, then W U B R G, so
	// the coloured mana survives for the next spell. It is the right
	// instinct for almost every cast.
	SpendPreserveColors ManaSpendStrategy = iota

	// SpendDistinctColors pays the generic half with a colour it has
	// not spent yet whenever it can. Converge (CR 702.86) and
	// sunburst (CR 702.44) count the COLOURS spent, so the ordinary
	// instinct is exactly backwards for them: a Painful Truths paid
	// out of a five-colour pool should spend five colours, not three
	// colourless and a Swamp.
	//
	// Chosen by the SPELL — effects.Spec.WantsDistinctColors, read
	// through CatalogWantsDistinctColors — and not by the player, so
	// CR 601.2h's "the player chooses which mana to spend" is still
	// the engine choosing well rather than a prompt.
	SpendDistinctColors
)

// SpendManaForWith is SpendManaFor with an explicit strategy, and
// returns the tokens it spent (#761). The tokens are the record every
// converge, sunburst and adamant card reads; they are returned rather
// than stamped anywhere here, because the pool does not know what it
// is paying for.
func (p *ManaPool) SpendManaForWith(cost ParsedCost, xValue int, ctx ManaSpendContext, strategy ManaSpendStrategy) ([]ManaToken, bool) {
	remaining, spent, ok := p.attemptSpend(cost, xValue, ctx, strategy)
	if !ok {
		return nil, false
	}
	*p = remaining
	return spent, true
}

// SpendMana attempts to deduct `cost` from the pool using only
// unrestricted tokens. On success, the pool is mutated and the method
// returns true. On failure, the pool is left unchanged. Caller
// usually calls CanPay first for a dry-run, then SpendMana to commit;
// this two-step shape keeps the "warn-and-proceed" sandbox path (S15
// sub-PR 3) cheap.
//
// Prefer SpendManaFor — same reasoning as CanPay vs CanPayFor.
func (p *ManaPool) SpendMana(cost ParsedCost, xValue int) bool {
	_, ok := p.SpendManaFor(cost, xValue, ManaSpendContext{})
	return ok
}

// SpendManaFor is SpendMana with a spend context. Pair it with the
// CanPayFor that gated the payment: calling CanPayFor with one
// context and SpendManaFor with another would let a restricted token
// pay for something it was checked against differently.
//
// Since #761 it returns the TOKENS it spent alongside the bool. The
// pool is where that fact exists and nowhere else — a moment later the
// tokens are gone and the Treasure that made one may be in a graveyard
// — so every payment path records what it gets back on the
// announcement's PaidCost.
func (p *ManaPool) SpendManaFor(cost ParsedCost, xValue int, ctx ManaSpendContext) ([]ManaToken, bool) {
	return p.SpendManaForWith(cost, xValue, ctx, SpendPreserveColors)
}

// attemptSpend is the shared core of CanPayFor + SpendManaFor.
// Returns (remaining-pool, spent-tokens, true) on success, (nil, nil,
// false) on failure. Operates on a copy so the caller's slice is
// never aliased.
//
// #761: the SPENT half is the point of the second return value. The
// pool is the only place the fact exists — a moment later the tokens
// are gone — and converge, sunburst, adamant and "if no mana was
// spent" all read it off the announcement's PaidCost.
func (p ManaPool) attemptSpend(cost ParsedCost, xValue int, ctx ManaSpendContext, strategy ManaSpendStrategy) (ManaPool, []ManaToken, bool) {
	// Copy so we can mark spends without disturbing the caller.
	work := make(ManaPool, len(p))
	copy(work, p)
	used := make([]bool, len(work))
	// One pass decides which tokens are legal here and in what
	// order the solver should prefer them. Both loops below walk
	// `order` rather than `work`, so the restriction-first
	// preference applies to colored requirements and to generic
	// alike, and pool insertion order still breaks ties — the
	// auto-tapper preview rendered alongside stays stable.
	order := spendOrder(work, ctx)

	// Step 1 — colored requirements first. Each requirement names
	// a set of allowed colors (monocolored = one entry, hybrid =
	// two). Pick the first available token whose color matches.
	// Phyrexian / numeric-alt requirements pay the color half in
	// S15 (life self-pay is S17).
	//
	// UNTOUCHED by the spend strategy (#761), deliberately: this is
	// the pass that decides whether a cost can be paid at all, and a
	// colour preference that could steer it into a dead end would
	// make payability depend on what the spell happens to read.
	//
	// A WIDENED slot (AnyMana — a spend grant's, #1589 / #1600) takes
	// a token of a colour it PRINTS here, like any other slot, and
	// otherwise joins the generic demand below, which any token pays.
	// That is the same answer to "payable?" as admitting any token
	// here, and a better answer to "with what?": the colour the card
	// asked for is spent when the pool has it, so a grant the player
	// "may" use never costs a Firespout its red mode.
	//
	// ADR 0118 §1: a requirement the first-match walk leaves unpaid is
	// given one more chance by reassignColored before the cost is
	// called unpayable. A payment the walk already makes is made with
	// exactly the same tokens as before.
	deferred := 0
	reqs := widenedLast(cost.Required)
	owner := newTokenOwners(len(work))
	for ri, req := range reqs {
		idx := firstPrintedMatch(work, used, order, req)
		if idx < 0 && !req.AnyMana {
			idx = reassignColored(work, used, order, owner, reqs, ri)
		}
		if idx < 0 {
			if req.AnyMana {
				deferred++
				continue
			}
			return nil, nil, false
		}
		used[idx] = true
		owner[idx] = ri
	}

	// Step 2 — generic requirement. Total = explicit Generic +
	// XSlots * xValue, plus the widened slots no printed colour paid
	// (#1600). The COLOUR ORDER is the strategy's one job:
	// colourless first to preserve colored for the next cast, or —
	// for a converge / sunburst spell — a colour this payment has
	// not spent yet, because those cards count colours rather than
	// mana. Every bucket is walked either way, so the answer to
	// "payable?" is the same under both.
	need := cost.Generic + cost.XSlots*xValue + deferred
	// #2170: mana that can't pay generic costs may still pay the
	// slots that only LOOK generic here — a widened coloured symbol
	// and a coloured symbol a cast permission folded into Generic. It
	// is spent on those first, because it can pay nothing else, and
	// never reaches the generic loops below.
	need -= spendNoGenericSlots(work, used, order, noGenericSlots(cost, deferred))
	if need > 0 && strategy == SpendDistinctColors {
		// One token at a time, because the question is about the SET
		// of colours: a bucket loop that emptied {W} before touching
		// {U} would pay a two-generic cost out of two Plains and
		// converge for one.
		spentColors := make(map[string]bool, 5)
		for i, tok := range work {
			if used[i] && isColorSymbol(tok.Color) {
				spentColors[tok.Color] = true
			}
		}
		for need > 0 {
			idx := pickDistinctColor(work, used, order, spentColors)
			if idx < 0 {
				return nil, nil, false
			}
			used[idx] = true
			if isColorSymbol(work[idx].Color) {
				spentColors[work[idx].Color] = true
			}
			need--
		}
	} else if need > 0 {
		for _, color := range []string{"C", "W", "U", "B", "R", "G"} {
			if need == 0 {
				break
			}
			for _, i := range order {
				if need == 0 {
					break
				}
				if used[i] {
					continue
				}
				if work[i].Color != color || work[i].noGeneric() {
					continue
				}
				used[i] = true
				need--
			}
		}
		if need > 0 {
			return nil, nil, false
		}
	}

	// Build remaining slice from un-used tokens, preserving order.
	// Tokens `order` excluded as unspendable were never marked used
	// and survive here, which is what has to happen: refusing to
	// spend restricted mana on this cast must not delete it.
	out := make(ManaPool, 0, len(work))
	var spent []ManaToken
	for i, tok := range work {
		if used[i] {
			spent = append(spent, tok)
			continue
		}
		out = append(out, tok)
	}
	if len(out) == 0 {
		return nil, spent, true
	}
	return out, spent, true
}

// noGenericSlots is how many of a cost's "generic" demand a mana that
// can't pay generic costs may still pay (#2170): the widened coloured
// symbols that no printed colour paid (`deferred`) — whether the grant
// is a player's static or a cast permission's (#1928). They are
// coloured symbols in the rules (CR 107.4e, 609.4b), only represented
// as generic demand here.
func noGenericSlots(_ ParsedCost, deferred int) int {
	return deferred
}

// spendNoGenericSlots marks up to n unused no-generic tokens (in
// `order`) as spent and returns how many it took. Shared by the pool
// solver and MissingFor so the two cannot disagree.
func spendNoGenericSlots(work ManaPool, used []bool, order []int, n int) int {
	taken := 0
	for _, i := range order {
		if taken == n {
			break
		}
		if used[i] || !work[i].noGeneric() {
			continue
		}
		used[i] = true
		taken++
	}
	return taken
}

// pickDistinctColor chooses the next token to pay one generic mana
// under SpendDistinctColors (#761), or -1 when nothing spendable is
// left. Three tiers, in order:
//
//  1. a COLOUR this payment has not spent yet — the whole point;
//  2. colourless, which never widens the count and is the cheapest
//     thing to give up;
//  3. a colour already spent, which is the ordinary "any mana will
//     do" case.
//
// Within a tier it walks `order`, the restriction-first sequence
// every other pass uses, so the pick is deterministic and a
// restricted token is still burned before an unrestricted one.
//
// It always returns a token when one is available, in every tier, so
// this strategy can never fail a payment the default would have made:
// generic mana accepts any colour, so payability is a question of
// COUNT and nothing else.
func pickDistinctColor(work ManaPool, used []bool, order []int, spentColors map[string]bool) int {
	fresh, colorless, repeat := -1, -1, -1
	for _, i := range order {
		if used[i] {
			continue
		}
		if work[i].noGeneric() {
			continue
		}
		switch {
		case isColorSymbol(work[i].Color) && !spentColors[work[i].Color]:
			if fresh < 0 {
				fresh = i
			}
		case work[i].Color == "C":
			if colorless < 0 {
				colorless = i
			}
		default:
			if repeat < 0 {
				repeat = i
			}
		}
	}
	for _, idx := range []int{fresh, colorless, repeat} {
		if idx >= 0 {
			return idx
		}
	}
	return -1
}

// spendOrder returns the indices of the tokens in `pool` that `ctx`
// permits, most-restricted first and pool order within a tier.
// Tokens the context does not permit are omitted entirely — the
// solver never sees them, so they can neither pay nor be consumed.
func spendOrder(pool ManaPool, ctx ManaSpendContext) []int {
	order := make([]int, 0, len(pool))
	for i, tok := range pool {
		if !ctx.allowsToken(tok) {
			continue
		}
		order = append(order, i)
	}
	// Stable insertion sort by descending restriction count. The
	// pool is a handful of entries in every real game, and an
	// insertion sort keeps the equal-count ordering identical to
	// pool insertion order without pulling in sort.SliceStable's
	// reflection.
	for i := 1; i < len(order); i++ {
		for j := i; j > 0; j-- {
			if len(pool[order[j]].Restrictions) <= len(pool[order[j-1]].Restrictions) {
				break
			}
			order[j], order[j-1] = order[j-1], order[j]
		}
	}
	return order
}

// widenedLast returns reqs with every AnyMana slot (#1589) moved
// behind the slots that still name their colours, order otherwise
// kept. The pool solver pays requirements greedily, first match wins,
// so a slot that admits any mana must choose last or it can take the
// one token a narrower slot behind it needed — a Dismember under an
// any-colour grant, taxed {W} by a cost modifier, would spend the
// only Plains on a {B/P} any Swamp could have paid. Returns reqs
// itself, unallocated, when nothing is widened — every cost with no
// spend grant.
func widenedLast(reqs []ColorRequirement) []ColorRequirement {
	widened := -1
	for i, r := range reqs {
		if r.AnyMana {
			widened = i
			break
		}
	}
	if widened < 0 {
		return reqs
	}
	out := make([]ColorRequirement, 0, len(reqs))
	for _, r := range reqs {
		if !r.AnyMana {
			out = append(out, r)
		}
	}
	for _, r := range reqs {
		if r.AnyMana {
			out = append(out, r)
		}
	}
	return out
}

// matchColor reports whether `color` is in `options`. Empty options
// means "any color" — never produced by ParseCost today, but cheap
// to handle defensively for future requirement shapes.
func matchColor(color string, options []string) bool {
	if len(options) == 0 {
		return true
	}
	for _, o := range options {
		if o == color {
			return true
		}
	}
	return false
}

// Missing returns the list of cost symbols `pool` cannot cover, in
// the order they appear in the cost. Returns nil when the cost is
// fully payable (CanPay would return true). Used by the strict-mode
// gate (S15 sub-PR 3) to populate the structured error frame the
// client renders into the override-toast affordance.
func (p ManaPool) Missing(cost ParsedCost, xValue int) []string {
	return p.MissingFor(cost, xValue, ManaSpendContext{})
}

// MissingFor is Missing with a spend context, so the breakdown a
// player sees matches the payment the engine would actually attempt.
// Without it a pool holding two Ancient Ziggurat {G} would report
// "missing nothing" for a Lightning Bolt the engine then refuses.
func (p ManaPool) MissingFor(cost ParsedCost, xValue int, ctx ManaSpendContext) []string {
	if p.CanPayFor(cost, xValue, ctx) {
		return nil
	}
	work := make(ManaPool, len(p))
	copy(work, p)
	used := make([]bool, len(work))
	order := spendOrder(work, ctx)
	var missing []string

	// attemptSpend's two passes, mirrored: a widened slot no printed
	// colour paid waits for the generic pass (#1600).
	var deferred []ColorRequirement
	reqs := widenedLast(cost.Required)
	owner := newTokenOwners(len(work))
	for ri, req := range reqs {
		idx := firstPrintedMatch(work, used, order, req)
		if idx < 0 && !req.AnyMana {
			idx = reassignColored(work, used, order, owner, reqs, ri)
		}
		if idx < 0 {
			if req.AnyMana {
				deferred = append(deferred, req)
				continue
			}
			missing = append(missing, "{"+formatRequirement(req)+"}")
			continue
		}
		used[idx] = true
		owner[idx] = ri
	}

	generic := cost.Generic + cost.XSlots*xValue
	need := len(deferred) + generic
	need -= spendNoGenericSlots(work, used, order, noGenericSlots(cost, len(deferred)))
	for _, color := range []string{"C", "W", "U", "B", "R", "G"} {
		if need == 0 {
			break
		}
		for _, i := range order {
			if need == 0 {
				break
			}
			if used[i] || work[i].Color != color || work[i].noGeneric() {
				continue
			}
			used[i] = true
			need--
		}
	}
	// The widened slots are paid before the generic, so what is still
	// owed is the generic and then, past it, the widened slots from the
	// back — named by their printed symbol, as they were before #1600.
	if short := need - generic; short > 0 {
		for _, req := range deferred[len(deferred)-short:] {
			missing = append(missing, "{"+formatRequirement(req)+"}")
		}
		need = generic
	}
	for need > 0 {
		missing = append(missing, "{1}")
		need--
	}
	return missing
}

// firstPrintedMatch is the token a coloured requirement takes in the
// pool solver's first pass: the first unused token in `order` whose
// colour the requirement PRINTS (matchColor over Options), or -1. A
// widened slot (AnyMana) is asked the same question here; what it
// admits beyond its printed colours is paid in the generic pass
// instead (#1600), so the two passes together admit exactly what
// ColorRequirement.Admits does.
func firstPrintedMatch(work ManaPool, used []bool, order []int, req ColorRequirement) int {
	for _, i := range order {
		if used[i] {
			continue
		}
		if matchColor(work[i].Color, req.Options) {
			return i
		}
	}
	return -1
}

// newTokenOwners is the colored pass's booking sheet: for each token,
// the index of the requirement it was booked for, or -1.
func newTokenOwners(n int) []int {
	owner := make([]int, n)
	for i := range owner {
		owner[i] = -1
	}
	return owner
}

// reassignColored finds a token for requirement `ri` when no unused
// token prints its colour, by moving a token already booked for an
// earlier requirement onto another token that requirement also
// prints: an augmenting path, so the colored pass pays every
// requirement whenever some assignment of tokens to requirements
// could (CR 601.2h: the player chooses which mana pays which symbol,
// and would choose an assignment that works). It returns the token
// now free for `ri`, with every booking along the path already moved,
// or -1 with every booking left as it was. The caller books the
// returned token for `ri`.
//
// It is asked only after firstPrintedMatch has failed, so a payment
// the first-match walk could already make is made with the same
// tokens: this repairs a refusal and never changes a success. ADR 0118
// §1's top-up is the case it exists for. A {G} that floated before the
// auto-tapper ran sits in front of the plan's own mana, so the walk
// alone pays {G/W}{G} out of [{G}, {W}] by spending the {G} on the
// hybrid and then finds no {G} for the second symbol, with lands
// already tapped for a cast it refuses.
//
// Widened slots (AnyMana) never come here: they are walked last and
// defer to the generic pass rather than fail, so no booking this moves
// can belong to one.
func reassignColored(work ManaPool, used []bool, order []int, owner []int, reqs []ColorRequirement, ri int) int {
	visited := make([]bool, len(work))
	var augment func(r int) int
	augment = func(r int) int {
		for _, i := range order {
			if visited[i] || !matchColor(work[i].Color, reqs[r].Options) {
				continue
			}
			visited[i] = true
			if !used[i] {
				return i
			}
			prev := owner[i]
			if prev < 0 || reqs[prev].AnyMana {
				continue
			}
			if j := augment(prev); j >= 0 {
				used[j] = true
				owner[j] = prev
				return i
			}
		}
		return -1
	}
	return augment(ri)
}

// manaSpentEvent builds the EventManaSpent breadcrumb for a payment
// (#761). Before this the event carried an Actor and a Source and
// nothing about the mana — so the log said "mana was spent", which is
// the one thing everybody watching already knew. It now carries the
// amount and the distinct colours, in WUBRG order, which is what a
// player watching a converge spell resolve actually wants to see.
//
// An empty payment still emits: "this cast cost nothing" is a fact,
// and a reader that had to infer it from a missing event would be
// making exactly the "no record means nothing was spent" mistake
// PaidCost.OnPaper exists to prevent.
func manaSpentEvent(actor, source uuid.UUID, spent []ManaToken) Event {
	rec := PaidCost{Mana: spent}
	return Event{
		Kind:   EventManaSpent,
		Actor:  actor,
		Source: source,
		Amount: len(spent),
		Colors: rec.ColorsSpent(),
	}
}

// emptyAllManaPoolsLocked clears every seated player's mana pool
// and emits one EventManaPoolEmptied per affected player. Called
// from runStepEntryHooksLocked at every step boundary (CR 106.4).
// What a player is allowed to keep is mana_keep.go's (#2166).
// No-op for players whose pool is already empty so step-cycle
// chatter stays quiet. Caller must hold g.mu.
func (g *Game) emptyAllManaPoolsLocked() {
	for _, p := range g.Seats {
		if len(p.ManaPool) == 0 {
			continue
		}
		n := g.sweepManaPoolLocked(p)
		if n == 0 {
			continue
		}
		g.EmitEvent(Event{
			Kind:   EventManaPoolEmptied,
			Actor:  p.ID,
			Amount: n,
		})
	}
}

// formatRequirement spells one requirement back the way it was
// printed, for the missing-symbols breakdown: "W", "W/U", "W/P",
// "W/U/P". The Phyrexian tail is printed because a player looking at
// "missing {W/U/P}" can see the payment the engine did not take —
// before #787 that symbol read as a plain "{W/U}" and the life option
// vanished from the message.
func formatRequirement(req ColorRequirement) string {
	if len(req.Options) == 0 {
		return "?"
	}
	out := req.Options[0]
	for _, o := range req.Options[1:] {
		out += "/" + o
	}
	if req.Phyrexian {
		out += "/P"
	}
	return out
}
