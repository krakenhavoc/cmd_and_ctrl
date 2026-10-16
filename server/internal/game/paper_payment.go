package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// paper_payment.go — #3855 (ADR 0068 amendment 2026-10-15, "What was
// paid on paper"): a waived payment that something READS asks the
// caster what they paid.
//
// Since #3844 a waived charge (a permissive cast or activation whose
// pool cannot cover the cost, or a strict-mode Cast anyway) spends what
// the pool can pay and waives only the shortfall. The shortfall was
// real mana the engine never saw, so a reader of the record — sunburst,
// converge, adamant, "the amount of mana spent", "if no mana was
// spent" — got a lower bound or a guess.
//
// The answer is to ASK, but only when the answer matters:
//
//   - The shortfall is zero: nothing was waived, nothing to ask.
//   - Nothing reads the payment: the cast stays one click, and the
//     record keeps the shortfall's AMOUNT (PaidCost.PaperMana): the
//     one-click path asserts the rest was paid outside the pool.
//   - Something reads it: the cast is refused with a
//     *PaperPaymentRequiredError naming the shortfall slot by slot, and
//     the client re-sends it with a PaperPayment — either the mana paid
//     outside the pool, which must be a legal payment of exactly the
//     shortfall, or Free, "an effect on paper let me cast it without
//     paying its mana cost". Either way the record becomes exact.
//
// The refusal comes before anything is paid, so a re-send starts from
// the same board.
//
// "Is it read" is a DECLARATION on the card (effects.Spec.ReadsManaSpent),
// enforced by a source scan in the effects package so a new reader
// cannot forget it.

// ManaSpentReads is what a card declares it reads of the mana spent to
// cast a spell or activate an ability (#3855). A bit set: a card can
// read its own payment and watch other casts.
type ManaSpentReads uint16

const (
	// ReadsOwnPayment: the card reads the mana spent to cast IT — at
	// resolution, from its own triggers, or as it enters (sunburst,
	// converge, adamant, Memory Deluge, Freestrider Commando).
	ReadsOwnPayment ManaSpentReads = 1 << iota
	// ReadsOwnActivations: an activated ability of the card reads the
	// mana spent to activate it (Protective Sphere).
	ReadsOwnActivations
	// ReadsYourCasts: a permanent that reads the payment of spells its
	// controller casts (Coin of Mastery, Lux Artillery, Satoru).
	ReadsYourCasts
	// ReadsOpponentsCasts: a permanent that reads the payment of spells
	// its controller's opponents cast (Boromir).
	ReadsOpponentsCasts
	// ReadsSourceKinds: the reading counts mana by the kind of source it
	// came from in a way a paper source can satisfy — today only "mana
	// from an artifact source" (Coin of Mastery). The picker then asks,
	// per mana, whether an artifact made it.
	ReadsSourceKinds
	// ReadsPoolManaOnly: the card reads only mana a paper payment can
	// never be — a Treasure's, a spend rider's token. The declaration
	// changes nothing it reads, so it never asks (Path of Ancestry,
	// Hired Hexblade, Kalain).
	ReadsPoolManaOnly
	// ReadsAnotherSpellsAmount: the card reads the AMOUNT spent on a
	// spell it does not watch the cast of — a counterspell reading its
	// target (Mana Sculpt). It cannot ask, since nothing at the cast
	// knows it is coming; the amount is exact without asking, because
	// every waived record carries its shortfall (PaidCost.PaperMana).
	ReadsAnotherSpellsAmount
)

// ReadsEveryCast is a permanent that watches every spell cast (Vexing
// Bauble, Roiling Vortex).
const ReadsEveryCast = ReadsYourCasts | ReadsOpponentsCasts

// CatalogReadsManaSpent is the catalog's ManaSpentReads for a card key.
// Nil hook ⇒ nothing reads anything, and every waived cast stays one
// click — what the game package's own tests expect.
var CatalogReadsManaSpent func(oracleID string) ManaSpentReads

func catalogReadsManaSpent(key string) ManaSpentReads {
	if CatalogReadsManaSpent == nil || key == "" {
		return 0
	}
	return CatalogReadsManaSpent(key)
}

// PaperMana is one mana a caster declares they paid outside the pool:
// its symbol (W U B R G or C) and, when a reader asks, whether an
// artifact made it. Count, when above one, makes the entry that many
// identical mana (#3904), so a declaration of a long shortfall stays
// small; 0 and 1 both mean one.
type PaperMana struct {
	Color    string `json:"color"`
	Artifact bool   `json:"artifact,omitempty"`
	Count    int    `json:"count,omitempty"`
	// X says this mana paid for the {X} part of the cost (#3941). The
	// picker sets it on the mana it prefilled from a slot marked X, so
	// the caster's own allocation (CR 601.2h) is what the record keeps.
	// Optional: a declaration with no X marks has the engine assign the
	// mana to the slots, giving the X part the colours the card reads
	// first.
	X bool `json:"x,omitempty"`
}

// PaperPayment is the caster's answer to a PaperPaymentRequiredError
// (#3855): either the mana they paid outside the pool, or Free — an
// effect the engine does not run let them cast it without paying its
// mana cost. Rides cast_spell and activate_ability as `paper_payment`.
type PaperPayment struct {
	Free bool        `json:"free,omitempty"`
	Mana []PaperMana `json:"mana,omitempty"`
}

// PaperSlot is one mana the pool left unpaid: its symbol as printed
// ("{R}", "{W/U}", "{1}") and the symbols that may pay it. Count, when
// above one, is a run of that many identical mana named once (#3904):
// a run longer than maxListedRun is one slot rather than one per mana,
// and is answered by one PaperMana with the same count.
type PaperSlot struct {
	Symbol  string   `json:"symbol"`
	Options []string `json:"options"`
	Count   int      `json:"count,omitempty"`
	// X marks a slot of the {X} part of the cost (#3941), which a
	// "spend only <colours> mana on X" clause restricts. The picker
	// shows these apart from the rest and marks the mana it declares for
	// them (PaperMana.X).
	X bool `json:"x,omitempty"`
}

// ErrPaperPaymentRequired is the sentinel a *PaperPaymentRequiredError
// wraps.
var ErrPaperPaymentRequired = errors.New("game: say what was paid outside the mana pool")

// ErrInvalidPaperPayment refuses a declaration that is not a legal
// payment of exactly what the pool left unpaid. Nothing is paid.
var ErrInvalidPaperPayment = fmt.Errorf("%w: the declared mana does not pay what the pool left unpaid", ErrInvalidParam)

// PaperPaymentRequiredError refuses a waived payment that something
// reads until the caster says what they paid (#3855). Nothing has been
// paid. Slots is the shortfall, one entry per mana (a long run is one
// entry with its count, #3904), coloured requirements first and then
// the generic; the client prefills each slot with its first option.
type PaperPaymentRequiredError struct {
	Slots []PaperSlot
	// ArtifactSources asks, per mana, whether an artifact made it
	// (ReadsSourceKinds).
	ArtifactSources bool
}

func (e *PaperPaymentRequiredError) Error() string { return ErrPaperPaymentRequired.Error() }

// Unwrap lets errors.Is(err, ErrPaperPaymentRequired) match.
func (e *PaperPaymentRequiredError) Unwrap() error { return ErrPaperPaymentRequired }

// paperSymbols is every symbol a generic slot admits, in WUBRG then C
// order — the order the picker lists them.
var paperSymbols = []string{"W", "U", "B", "R", "G", "C"}

// paperShortfall is what a waived payment left unpaid: the coloured
// requirements no pool mana could take, as runs (#3904), and the
// any-mana remainder.
type paperShortfall struct {
	reqs    []reqRun
	generic int
	// favor is the colours the X part takes first when a declaration
	// leaves the allocation to the engine (#3941).
	favor []string
}

// shortfallFor is what SpendWhatPaysForWith would leave unpaid, read
// without spending.
func (p ManaPool) shortfallFor(cost ParsedCost, xValue int, ctx ManaSpendContext, strategy ManaSpendStrategy) paperShortfall {
	_, _, unpaid, need := p.allocateRest(cost, xValue, ctx, strategy)
	return paperShortfall{reqs: unpaid, generic: need, favor: ctx.favoredOnX()}
}

func (s paperShortfall) amount() int { return satAddMana(reqRunsTotal(s.reqs), s.generic) }

// slots is the shortfall one mana at a time, for the picker, except
// that a run longer than maxListedRun is one slot with its count
// (#3904): the shortfall can carry an announced X, and a picker of X
// slots is neither sendable nor answerable.
func (s paperShortfall) slots() []PaperSlot {
	out := make([]PaperSlot, 0, min(s.amount(), maxPreallocSlots))
	for _, r := range s.reqs {
		opts := append([]string(nil), r.req.Options...)
		if len(opts) == 0 {
			opts = append([]string(nil), paperSymbols...)
		}
		out = appendPaperSlots(out, PaperSlot{Symbol: "{" + formatRequirement(r.req) + "}", Options: opts, X: r.x}, r.n)
	}
	if s.generic > 0 {
		out = appendPaperSlots(out, PaperSlot{Symbol: "{1}", Options: append([]string(nil), paperSymbols...)}, s.generic)
	}
	return out
}

// appendPaperSlots appends n of slot: one per mana up to maxListedRun,
// and past it one slot carrying the count.
func appendPaperSlots(out []PaperSlot, slot PaperSlot, n int) []PaperSlot {
	if n > maxListedRun {
		slot.Count = n
		return append(out, slot)
	}
	for range n {
		one := slot
		one.Options = append([]string(nil), slot.Options...)
		out = append(out, one)
	}
	return out
}

// String is the shortfall as a cost, generic first: "{4}{G}{G}". A run
// longer than maxListedRun is written once with its count, "{B}×500000".
func (s paperShortfall) String() string {
	var b strings.Builder
	if s.generic > 0 {
		b.WriteString("{" + strconv.Itoa(s.generic) + "}")
	}
	for _, r := range coalesced(s.reqs) {
		writeSymbolRun(&b, "{"+formatRequirement(r.req)+"}", r.n)
	}
	return b.String()
}

// tokensFor validates a declaration against the shortfall and returns
// the mana it declares, as tokens with no source (they came from
// outside the engine) and the artifact kind when declared. Free returns
// none. A declaration must be a legal payment of EXACTLY the shortfall:
// as many mana as it is short, each coloured requirement paid by a
// symbol it admits (a hybrid by either half, never by a third colour),
// and the rest paying generic.
//
// A counted entry (PaperMana.Count) is one run of that many mana. The
// counts are checked against the shortfall as they are read, so no
// count, however large, is summed past it or allocated by (#3904).
func (s paperShortfall) tokensFor(d PaperPayment) ([]ManaToken, error) {
	if d.Free {
		if len(d.Mana) > 0 {
			return nil, ErrInvalidPaperPayment
		}
		return nil, nil
	}
	amount := s.amount()
	if len(d.Mana) > amount {
		return nil, ErrInvalidPaperPayment
	}
	pool := make(ManaPool, len(d.Mana))
	order := make([]int, len(d.Mana))
	total, marked := 0, 0
	for i, m := range d.Mana {
		if !isColorSymbol(m.Color) && m.Color != "C" {
			return nil, ErrInvalidPaperPayment
		}
		n := max(m.Count, 1)
		if m.Count < 0 || n > amount-total {
			return nil, ErrInvalidPaperPayment
		}
		total += n
		if m.X {
			marked += n
		}
		pool[i] = ManaToken{Color: m.Color}.withN(n)
		if m.Artifact {
			pool[i].SourceKinds = ManaSourceArtifact
		}
		order[i] = i
	}
	if total != amount {
		return nil, ErrInvalidPaperPayment
	}
	if marked > 0 {
		return s.markedTokens(d, pool)
	}
	// No marks (#3941): the caster named the mana and left the matching
	// to the engine, which gives the X part the colours the card reads
	// first (the same choice the pool's payment makes).
	used := make([]int, len(pool))
	_, unpaid, xUsed := colouredPassX(pool, used, order, s.reqs, s.favor)
	if len(unpaid) > 0 {
		return nil, ErrInvalidPaperPayment
	}
	out := make([]ManaToken, 0, len(pool)+1)
	for i, t := range pool {
		xu := 0
		if xUsed != nil {
			xu = xUsed[i]
		}
		if xu > 0 {
			x := t.withN(xu)
			x.ForX = true
			out = append(out, x)
		}
		if rest := t.N() - xu; rest > 0 {
			out = append(out, t.withN(rest))
		}
	}
	return out, nil
}

// markedTokens is tokensFor for a declaration that says which mana paid
// the X part (#3941): the marked mana must be exactly the X slots' worth
// and pay them, and the rest must pay everything else, each against its
// own requirements. The marked tokens are recorded ForX.
func (s paperShortfall) markedTokens(d PaperPayment, pool ManaPool) ([]ManaToken, error) {
	var xReqs, otherReqs []reqRun
	for _, r := range s.reqs {
		if r.x {
			xReqs = append(xReqs, r)
		} else {
			otherReqs = append(otherReqs, r)
		}
	}
	var xPool, otherPool ManaPool
	for i, m := range d.Mana {
		if m.X {
			xPool = append(xPool, pool[i])
		} else {
			otherPool = append(otherPool, pool[i])
		}
	}
	if len(xReqs) == 0 || TotalMana(xPool) != reqRunsTotal(xReqs) {
		return nil, ErrInvalidPaperPayment
	}
	pays := func(p ManaPool, reqs []reqRun) bool {
		order := make([]int, len(p))
		for i := range order {
			order[i] = i
		}
		_, unpaid := colouredPass(p, make([]int, len(p)), order, reqs)
		return len(unpaid) == 0
	}
	if !pays(xPool, xReqs) || !pays(otherPool, otherReqs) {
		return nil, ErrInvalidPaperPayment
	}
	out := make([]ManaToken, len(pool))
	for i, t := range pool {
		t.ForX = d.Mana[i].X
		out[i] = t
	}
	return out, nil
}

// paperSymbolsOf is a declaration as a cost string for the log, in the
// order declared: "{C}{C}{R}". A counted entry longer than
// maxListedRun is written once with its count, "{B}×500000".
func paperSymbolsOf(d PaperPayment) string {
	var b strings.Builder
	for _, m := range d.Mana {
		writeSymbolRun(&b, "{"+m.Color+"}", max(m.Count, 1))
	}
	return b.String()
}

// paperNote is what the cast's log line says about a waived payment.
type paperNote struct {
	// Declared: the caster answered a PaperPayment.
	Declared bool
	// Free: they declared they paid no mana cost.
	Free bool
	// Symbols: the declared mana, or the undeclared shortfall.
	Symbols string
}

// castPaymentReadLocked reports whether anything reads the mana spent
// to cast `card` by `caster` (#3855): the spell itself, or a permanent
// watching casts — its controller's or its opponents'. The second
// result asks for the source kinds (ReadsSourceKinds).
func (g *Game) castPaymentReadLocked(caster uuid.UUID, card Card) (read, kinds bool) {
	own := catalogReadsManaSpent(CatalogKey(card))
	if own&ReadsOwnPayment != 0 {
		read = true
		kinds = own&ReadsSourceKinds != 0
	}
	if g.Battlefield == nil {
		return read, kinds
	}
	for i := range g.Battlefield.Cards {
		perm := &g.Battlefield.Cards[i]
		r := catalogReadsManaSpent(CatalogKeyOf(perm))
		watch := ReadsOpponentsCasts
		if perm.Controller == caster {
			watch = ReadsYourCasts
		}
		if r&watch == 0 {
			continue
		}
		read = true
		if r&ReadsSourceKinds != 0 {
			kinds = true
		}
	}
	return read, kinds
}

// abilityPaymentRead reports whether an activated ability of
// `source` reads the mana spent to activate it.
func abilityPaymentRead(source Card) (read, kinds bool) {
	r := catalogReadsManaSpent(CatalogKey(source))
	return r&ReadsOwnActivations != 0, r&ReadsOwnActivations != 0 && r&ReadsSourceKinds != 0
}

// settlePaperPaymentLocked is the shared head of a waived payment
// (#3855), run BEFORE anything is paid. It returns the declared tokens
// to append to the record (nil for Free or no declaration) and the
// log note, or a refusal: *PaperPaymentRequiredError when something
// reads the payment and nothing was declared, ErrInvalidPaperPayment
// for a declaration that does not pay the shortfall.
func settlePaperPayment(short paperShortfall, decl *PaperPayment, read, kinds bool) ([]ManaToken, paperNote, error) {
	if short.amount() == 0 {
		return nil, paperNote{}, nil
	}
	if decl == nil {
		if read {
			return nil, paperNote{}, &PaperPaymentRequiredError{Slots: short.slots(), ArtifactSources: kinds}
		}
		return nil, paperNote{Symbols: short.String()}, nil
	}
	tokens, err := short.tokensFor(*decl)
	if err != nil {
		return nil, paperNote{}, err
	}
	if decl.Free {
		return nil, paperNote{Declared: true, Free: true}, nil
	}
	return tokens, paperNote{Declared: true, Symbols: paperSymbolsOf(*decl)}, nil
}

// paperLogSymbols is the mana EventCast names as paid on paper: the
// declaration, or a one-click Cast anyway's shortfall. Empty for a
// permissive cast nobody asked about (ADR 0118 owner decision 2: that
// seat may be paying on paper, and the log stays quiet) and for a cast
// declared free, which the log words as "without paying its mana cost".
func paperLogSymbols(params CastSpellParams, note paperNote) string {
	if note.Free || (!note.Declared && !params.ForceCast) {
		return ""
	}
	return note.Symbols
}

// declaredSymbols is the declared mana an activation's cost warning
// names: only a declaration is logged there, since a permissive
// activation nobody asked about stays quiet.
func declaredSymbols(note paperNote) string {
	if !note.Declared || note.Free {
		return ""
	}
	return note.Symbols
}

// recordPaperPayment completes a waived payment's record once the pool
// has paid its share: the declared mana joins the tokens and the record
// is exact; Free leaves the pool's tokens as the whole payment, also
// exact; no declaration keeps OnPaper and records the shortfall's
// amount as paid outside the pool.
func recordPaperPayment(paid *PaidCost, short paperShortfall, declared []ManaToken, note paperNote) {
	if short.amount() == 0 {
		return
	}
	if note.Declared {
		paid.Mana = append(paid.Mana, declared...)
		return
	}
	paid.OnPaper = true
	paid.PaperMana = short.amount()
}
