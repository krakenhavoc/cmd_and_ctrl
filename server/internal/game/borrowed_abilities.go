package game

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// borrowed_abilities.go — "has all activated abilities of …" (ADR 0155,
// #1557). CR 113.10 and CR 613.1f: an effect that gives an object every
// activated ability ANOTHER object has right now — Necrotic Ooze, Marvin,
// Experiment Kraj, Manascape Refractor. ADR 0093 Decision 10 sent the
// case here: a grant names a fixed catalog bundle, and a borrow's lent set
// is "whatever these other objects have right now".
//
// # The shape
//
// A layer-6 static declares StaticAbility.Borrows (Decision 1). On every
// layer pass the borrow step, the LAST step of the layer-6 bucket
// (Decision 4), reads each lender's finished ability rows and writes one
// BorrowedAbility per lent ROW onto every recipient's Characteristic
// (Decision 2). An entry names the catalog row it came from — a card's own
// key or a grant bundle's key, and the row's index there — never a
// closure, so the Characteristic stays plain data and the stack item of
// an activated borrowed row needs no lender to resolve (CR 113.7a).
//
// The two row accessors (activatedAbilityRows and manaAbilityRows) read
// the entries as a FOURTH half after own, intrinsic and granted, so a
// borrow appearing never renumbers an existing row. Nothing else reads
// them: a borrow is not composed into CatalogAbilityKey, so a lender's
// triggers, statics and cost modifiers are never lent.
//
// # What "this creature" means
//
// The borrower. A row's Effect reads item.SourceCardID, which is the
// borrower, so {T} taps it, "sacrifice this creature" sacrifices it and
// its controller activates the ability (ADR 0155 Decision 6). The
// lender's printed name in a row's label is the borrower's name on the
// wire and in the log (CR 201.5b, BorrowedLabel).
//
// # Instance-carried rows
//
// The ADR's LentInstance kind is not written. An instance-carried row (a
// legacy token's Card.ActivatedAbilities) is a closure on the LENDER, and
// the row accessors hold the borrower alone, with no game to find the
// lender through. Such a row is lent as its catalog row when the lender's
// catalog identity (oracle ID or token key) hands the same row back at
// the same index — every token minted from the token catalog does — and
// is not lent otherwise, which is weaker than printed, never stronger.

// AbilityBorrow is a layer-6 static's "has all activated abilities of …"
// declaration (ADR 0155 Decision 1). The recipients are the static's
// AppliesTo, as for any grant; the lenders are read here on every pass.
type AbilityBorrow struct {
	// Lenders lists the objects whose activated abilities are lent, read
	// once per layer pass, after layers 1–5 have settled (control, name,
	// types). Catalog code; never stored.
	Lenders func(g *Game, source *Card) []*Card
	// Except leaves out a kind of row: BorrowExceptLoyalty (Scheming
	// Fence), BorrowExceptMana (Sharkey). Zero lends everything.
	Except BorrowExcept
	// EachOncePerTurn is "You may activate each of those abilities only
	// once each turn" (Mairsil, the Pretender; Locus of Enlightenment).
	EachOncePerTurn bool
	// Label is the printed clause, for the wire's granted_abilities line
	// and the borrow's identity.
	Label string
}

// BorrowExcept is a set of row kinds a borrow does not lend.
type BorrowExcept uint8

const (
	// BorrowExceptLoyalty is "except for loyalty abilities" (CR 606).
	BorrowExceptLoyalty BorrowExcept = 1 << iota
	// BorrowExceptMana is "activated abilities … that aren't mana
	// abilities" (CR 605.1a).
	BorrowExceptMana
)

// LentKind says which list a lent row is read from.
type LentKind uint8

const (
	// LentCatalog is a row of a catalog definition: Key names the list
	// (a card's own key or a grant bundle's GrantKey), Index the row.
	LentCatalog LentKind = iota
	// LentLand is a CR 305.6 intrinsic land mana ability: Color names it.
	LentLand
)

// BorrowSlotMana is a BorrowedAbility's Slot for a lent mana ability;
// AbilitySlotActivated is the other.
const BorrowSlotMana = "mana"

// BorrowedAbility is one lent row on a recipient (ADR 0155 Decision 2):
// layer-6 output, rebuilt on every pass, never composed into
// CatalogAbilityKey. Plain data, so the last-known-information copy of a
// Characteristic carries it.
type BorrowedAbility struct {
	// Slot is AbilitySlotActivated or BorrowSlotMana.
	Slot string
	// Kind is where the row is read from.
	Kind LentKind
	// Key is, for LentCatalog, the catalog key whose list holds the row:
	// a card's own key (face-aware) or a bundle's GrantKey.
	Key string
	// Index is the row's index in that list, declared and pre-gating.
	Index int
	// Color is, for LentLand, the CR 305.6 colour letter.
	Color string
	// Lender is the object the row was read off, and LenderName its
	// printed name: the wire's borrowed_from and CR 201.5b's rewrite. A
	// row a lender had itself borrowed keeps the object it came from.
	Lender     uuid.UUID
	LenderName string
	// Source is the borrowing static's source (the row's granted_by).
	Source uuid.UUID
	// Activator is the "you" of the borrowing static: its source's
	// controller when the pass wrote the entry.
	Activator uuid.UUID
	// Borrow is the borrowing static's identity, "<source>|<label>": what
	// EachOncePerTurn's tally and a mana-spend rider key on.
	Borrow string
	// EachOncePerTurn copies AbilityBorrow.EachOncePerTurn, which the
	// activation gate reads without the static.
	EachOncePerTurn bool
}

// sameRow reports whether two entries lend the same catalog row — CR
// 113.2c's "two instances of the same ability", which a ref's <n> counts.
func (b BorrowedAbility) sameRow(o BorrowedAbility) bool {
	return b.Kind == o.Kind && b.Key == o.Key && b.Index == o.Index && b.Color == o.Color
}

// borrowIdentity is a borrowing static's identity: its source and label.
func borrowIdentity(source uuid.UUID, label string) string {
	return source.String() + "|" + label
}

// BorrowLabelOf is the printed clause a BorrowedAbility.Borrow names.
func BorrowLabelOf(borrow string) string {
	if i := strings.IndexByte(borrow, '|'); i >= 0 {
		return borrow[i+1:]
	}
	return ""
}

// LentByBorrow reports whether a row was lent by a borrow whose printed
// clause is `label`: what a "spend mana as though it were mana of any
// color to activate those abilities" rider's Covers asks of
// ManaSpendContext.AbilityOrigin (ADR 0155 Decision 7). Two borrows with
// the same clause make the same promise, so the clause is identity
// enough.
func (o AbilityOrigin) LentByBorrow(label string) bool {
	return o.Borrowed && o.Lent != nil && BorrowLabelOf(o.Lent.Borrow) == label
}

// layeredBorrows returns the object's current lent rows, or nil for every
// object that has none.
func layeredBorrows(c *Card) []BorrowedAbility {
	if c.effective == nil {
		return nil
	}
	return c.effective.BorrowedAbilities
}

// HasBorrowedAbilities reports whether any row of the object is lent —
// the view's fast negative before it asks for origins.
func HasBorrowedAbilities(c *Card) bool { return len(layeredBorrows(c)) > 0 }

// --- refs (ADR 0155 Decision 5) ------------------------------------------

const abilityRefBorrow = "borrow:"

// BorrowedAbilityRef is the ref of a lent row: borrow:<key>:<i>:<n> for a
// catalog row, borrow:land:<colour>:<n> for an intrinsic land row, where
// <n> counts the earlier entries lending the same row.
func BorrowedAbilityRef(b BorrowedAbility, n int) string {
	if b.Kind == LentLand {
		return abilityRefBorrow + "land:" + b.Color + ":" + strconv.Itoa(n)
	}
	return abilityRefBorrow + b.Key + ":" + strconv.Itoa(b.Index) + ":" + strconv.Itoa(n)
}

// borrowedOrigin is the AbilityOrigin of a lent row; b points into the
// borrower's layered characteristic.
func borrowedOrigin(b *BorrowedAbility, n int) AbilityOrigin {
	return AbilityOrigin{
		Ref:       BorrowedAbilityRef(*b, n),
		GrantedBy: b.Source,
		Activator: b.Activator,
		Borrowed:  true,
		LentBy:    b.Lender,
		Lent:      b,
	}
}

// borrowOccurrence is how many earlier entries in `list` of the same slot
// lend the same row as list[i].
func borrowOccurrence(list []BorrowedAbility, i int) int {
	n := 0
	for j := 0; j < i; j++ {
		if list[j].Slot == list[i].Slot && list[j].sameRow(list[i]) {
			n++
		}
	}
	return n
}

// borrowedActivatedRows is the fourth half of activatedAbilityRows: each
// lent activated row, looked up in the catalog list its entry names. An
// entry whose row the catalog no longer has is skipped.
func borrowedActivatedRows(c *Card, wantOrigins bool) ([]ActivatedAbilityShape, AbilityOrigins) {
	list := layeredBorrows(c)
	if len(list) == 0 || CatalogActivatedAbilities == nil {
		return nil, nil
	}
	n := 0
	for i := range list {
		if list[i].Slot == AbilitySlotActivated {
			n++
		}
	}
	if n == 0 {
		return nil, nil
	}
	out := make([]ActivatedAbilityShape, 0, n)
	var origins AbilityOrigins
	if wantOrigins {
		origins = make(AbilityOrigins, 0, n)
	}
	// Lent rows come in runs from one lender, so the catalog list is
	// looked up once per run.
	var lastKey string
	var rows []ActivatedAbilityShape
	for i := range list {
		b := &list[i]
		if b.Slot != AbilitySlotActivated || b.Kind != LentCatalog {
			continue
		}
		if b.Key != lastKey || rows == nil {
			lastKey, rows = b.Key, CatalogActivatedAbilities(b.Key)
		}
		if b.Index < 0 || b.Index >= len(rows) {
			continue
		}
		out = append(out, rows[b.Index])
		if wantOrigins {
			origins = append(origins, borrowedOrigin(b, borrowOccurrence(list, i)))
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, origins
}

// borrowedManaRows is the fourth half of manaAbilityRows.
func borrowedManaRows(c *Card, wantOrigins bool) ([]ManaAbilityShape, AbilityOrigins) {
	list := layeredBorrows(c)
	if len(list) == 0 {
		return nil, nil
	}
	var out []ManaAbilityShape
	var origins AbilityOrigins
	for i := range list {
		b := &list[i]
		if b.Slot != BorrowSlotMana {
			continue
		}
		var row ManaAbilityShape
		switch b.Kind {
		case LentLand:
			r, ok := intrinsicLandRowFor(b.Color)
			if !ok {
				continue
			}
			row = r
		case LentCatalog:
			if CatalogManaAbilities == nil {
				continue
			}
			rows := CatalogManaAbilities(b.Key)
			if b.Index < 0 || b.Index >= len(rows) {
				continue
			}
			row = rows[b.Index]
		default:
			continue
		}
		out = append(out, row)
		if wantOrigins {
			origins = append(origins, borrowedOrigin(b, borrowOccurrence(list, i)))
		}
	}
	return out, origins
}

// intrinsicLandRowFor is the CR 305.6 row that adds `color`.
func intrinsicLandRowFor(color string) (ManaAbilityShape, bool) {
	for i, lt := range landTypeMana {
		if lt.Color == color {
			return intrinsicLandRows[i][0], true
		}
	}
	return ManaAbilityShape{}, false
}

// --- reading a lender (ADR 0155 Decision 3) ------------------------------

// lentRowsOf lists the rows `lender` lends right now, with only the row
// identity filled in. A lender on the battlefield is read from its
// finished layer-6 rows — grants, removals and rows it borrowed itself
// included — and one anywhere else from its catalog key (face-aware), its
// designation gate and its printed land types, through the same two
// accessors. A face-down lender lends nothing (CR 708.2a), and so does an
// uncatalogued one: the engine does not know its abilities.
func lentRowsOf(lender *Card, except BorrowExcept) []BorrowedAbility {
	if lender == nil || lender.FaceDown {
		return nil
	}
	var out []BorrowedAbility
	acts, ao := activatedAbilityRows(lender, true)
	for i := range acts {
		if except&BorrowExceptLoyalty != 0 && acts[i].Cost.Loyalty != nil {
			continue
		}
		if e, ok := lentEntryFor(lender, ao.At(i), AbilitySlotActivated, acts[i].Label); ok {
			out = append(out, e)
		}
	}
	if except&BorrowExceptMana == 0 {
		manas, mo := manaAbilityRows(lender, true)
		for i := range manas {
			if e, ok := lentEntryFor(lender, mo.At(i), BorrowSlotMana, manas[i].Label); ok {
				out = append(out, e)
			}
		}
	}
	return out
}

// lentEntryFor turns one of a lender's row origins into the entry that
// lends it, or false when the row has no catalog identity to name.
func lentEntryFor(lender *Card, o AbilityOrigin, slot, label string) (BorrowedAbility, bool) {
	if o.Borrowed && o.Lent != nil {
		// Re-lent as it is: the row, and the object it came from.
		return *o.Lent, true
	}
	e := BorrowedAbility{Slot: slot, Lender: lender.InstanceID, LenderName: lender.Name}
	switch {
	case o.Granted():
		_, i, _, ok := parseGrantAbilityRef(o.Ref)
		if !ok {
			return BorrowedAbility{}, false
		}
		e.Kind, e.Key, e.Index = LentCatalog, o.Grant, i
	case strings.HasPrefix(o.Ref, abilityRefLand):
		e.Kind, e.Color = LentLand, strings.TrimPrefix(o.Ref, abilityRefLand)
	case strings.HasPrefix(o.Ref, abilityRefOwn):
		i, err := strconv.Atoi(strings.TrimPrefix(o.Ref, abilityRefOwn))
		if err != nil || i < 0 {
			return BorrowedAbility{}, false
		}
		key, ok := ownRowKey(lender, slot, i, label)
		if !ok {
			return BorrowedAbility{}, false
		}
		e.Kind, e.Key, e.Index = LentCatalog, key, i
	default:
		return BorrowedAbility{}, false
	}
	return e, true
}

// ownRowKey is the catalog key whose list holds a lender's own row i. A
// catalog-read row is under the lender's own key. An instance-carried row
// is under its catalog identity only when that list hands the same row
// back at the same index (see the file comment).
func ownRowKey(lender *Card, slot string, i int, label string) (string, bool) {
	instance := len(lender.ActivatedAbilities) > 0
	if slot == BorrowSlotMana {
		instance = len(lender.ManaAbilities) > 0
	}
	// ADR 0160: under an exchanged text box the carried list is not
	// the lender's text, and the exchanged key holds every row.
	if !instance || lender.TextBox != nil {
		key := catalogKeyOf(lender)
		return key, key != ""
	}
	key := abilityCatalogKey(lender.OracleID, lender.TokenKey)
	if key == "" {
		return "", false
	}
	if slot == BorrowSlotMana {
		if CatalogManaAbilities == nil {
			return "", false
		}
		rows := CatalogManaAbilities(key)
		return key, i < len(rows) && rows[i].Label == label
	}
	if CatalogActivatedAbilities == nil {
		return "", false
	}
	rows := CatalogActivatedAbilities(key)
	return key, i < len(rows) && rows[i].Label == label
}

// --- the layer-6 borrow step (ADR 0155 Decision 4) -----------------------

// isBorrowEffect reports whether a layer-6 effect is a borrow.
func isBorrowEffect(eff ContinuousEffect) bool {
	sce, ok := eff.(staticContinuousEffect)
	return ok && sce.ability.Borrows != nil
}

// splitBorrows takes the borrows out of a layer-6 bucket, keeping both
// halves in timestamp order. The bucket comes back untouched, with no
// allocation, when it has none — every board but a handful.
func splitBorrows(bucket []ContinuousEffect) (rest, borrows []ContinuousEffect) {
	found := false
	for _, eff := range bucket {
		if isBorrowEffect(eff) {
			found = true
			break
		}
	}
	if !found {
		return bucket, nil
	}
	rest = make([]ContinuousEffect, 0, len(bucket))
	for _, eff := range bucket {
		if isBorrowEffect(eff) {
			borrows = append(borrows, eff)
		} else {
			rest = append(rest, eff)
		}
	}
	return rest, borrows
}

// pendingBorrow is one borrow of this pass with its lender set, read once.
type pendingBorrow struct {
	eff     staticContinuousEffect
	lenders []*Card
}

// applyBorrowsLocked is the borrow step: every borrow in the layer-6
// bucket, after every other layer-6 effect, ordered among themselves.
//
//  1. After everything else, because every ordinary layer-6 effect that
//     touches a lender changes what the borrow does (CR 613.8a): the
//     borrow reads the lender's finished rows.
//  2. Among borrows, B before A when B's recipients include one of A's
//     lenders; a cycle is CR 613.8b's loop, and timestamp order decides.
//  3. A borrow whose source has lost its abilities in this pass does
//     nothing, whatever the timestamps (ADR 0093 Decision 3).
//  4. A recipient that is not the source, under a layer-6 removal with a
//     LATER timestamp than the borrow, is skipped (CR 613.6): in
//     timestamp order that removal would have emptied what the borrow
//     wrote.
func (g *Game) applyBorrowsLocked(borrows []ContinuousEffect, st *layerPassState) {
	pending := make([]pendingBorrow, 0, len(borrows))
	for _, eff := range borrows {
		sce := eff.(staticContinuousEffect)
		var lenders []*Card
		if sce.ability.Borrows.Lenders != nil {
			lenders = sce.ability.Borrows.Lenders(g, sce.source)
		}
		pending = append(pending, pendingBorrow{eff: sce, lenders: lenders})
	}
	for _, i := range g.orderBorrowsLocked(pending) {
		g.applyOneBorrowLocked(pending[i], st)
	}
}

// orderBorrowsLocked is point 2 of applyBorrowsLocked: the order to apply
// the pending borrows in. A single borrow, or a set with no borrow
// feeding another, keeps timestamp order.
func (g *Game) orderBorrowsLocked(pending []pendingBorrow) []int {
	n := len(pending)
	order := make([]int, 0, n)
	if n == 1 {
		return append(order, 0)
	}
	// waits[a] lists the borrows a waits on: those whose recipients
	// include one of a's lenders.
	var waits [][]int
	for a := range pending {
		for b := range pending {
			if a == b {
				continue
			}
			for _, l := range pending[a].lenders {
				if l != nil && pending[b].eff.AppliesTo(l, g) {
					if waits == nil {
						waits = make([][]int, n)
					}
					waits[a] = append(waits[a], b)
					break
				}
			}
		}
	}
	done := make([]bool, n)
	for len(order) < n {
		pick, first := -1, -1
		for i := 0; i < n; i++ {
			if done[i] {
				continue
			}
			if first < 0 {
				first = i
			}
			ready := true
			if waits != nil {
				for _, w := range waits[i] {
					if !done[w] {
						ready = false
						break
					}
				}
			}
			if ready {
				pick = i
				break
			}
		}
		if pick < 0 {
			// CR 613.8b: a loop — timestamp order, the earliest left.
			pick = first
		}
		done[pick] = true
		order = append(order, pick)
	}
	return order
}

// applyOneBorrowLocked writes one borrow's lent rows onto its recipients.
// The lenders' rows are read once, before anything is written, so a
// recipient that is also a lender of the same borrow is not fed its own
// output in the same step.
func (g *Game) applyOneBorrowLocked(pb pendingBorrow, st *layerPassState) {
	eff := pb.eff
	src := effectSourceLocked(eff)
	if src != nil && src.HasLostAllAbilities() {
		return
	}
	var rows []BorrowedAbility
	for _, l := range pb.lenders {
		rows = append(rows, lentRowsOf(l, eff.ability.Borrows.Except)...)
	}
	if len(rows) == 0 {
		return
	}
	var from, you uuid.UUID
	if eff.source != nil {
		from, you = eff.source.InstanceID, eff.source.Controller
	}
	id := borrowIdentity(from, eff.ability.Borrows.Label)
	ts := eff.Timestamp()
	for i := range g.Battlefield.Cards {
		target := &g.Battlefield.Cards[i]
		if !eff.AppliesTo(target, g) {
			continue
		}
		if st.removedAfter(target, ts) {
			continue
		}
		for _, r := range rows {
			r.Source, r.Activator, r.Borrow = from, you, id
			r.EachOncePerTurn = eff.ability.Borrows.EachOncePerTurn
			target.effective.BorrowedAbilities = append(target.effective.BorrowedAbilities, r)
		}
	}
}

// --- the activation side -------------------------------------------------

// borrowedTallyKey is the per-turn tally key of one lent row on one
// borrower OBJECT, for EachOncePerTurn (ADR 0155 Decision 7): (borrower,
// borrow, row). Two identical lent rows share a key, which is the
// Mairsil ruling. Its "borrow:" prefix keeps it out of every per-object
// sum ActivatedThisTurn makes.
func (g *Game) borrowedTallyKeyLocked(borrower uuid.UUID, o AbilityOrigin) string {
	if !o.Borrowed || o.Lent == nil || !o.Lent.EachOncePerTurn {
		return ""
	}
	b := o.Lent
	row := b.Key + ":" + strconv.Itoa(b.Index)
	if b.Kind == LentLand {
		row = "land:" + b.Color
	}
	return abilityRefBorrow + g.objectTallyPrefixLocked(borrower) + b.Borrow + "|" + b.Slot + "|" + row
}

// BorrowedRowSpentLocked reports whether a lent row from an
// EachOncePerTurn borrow has already been activated by this borrower this
// turn. The one reader the activation paths refuse on, the enumerator
// drops the move for and the view greys the row by. False for every row
// that is not such a lent row.
//
// Caller must hold g.mu.
func (g *Game) BorrowedRowSpentLocked(borrower uuid.UUID, o AbilityOrigin) bool {
	key := g.borrowedTallyKeyLocked(borrower, o)
	return key != "" && g.Activations.Turn[key] > 0
}

// noteBorrowedActivationLocked records one activation of a lent
// once-each-turn row, in the per-turn half only. The key is taken before
// any cost is paid, for activationTallyKeyLocked's reason.
func (g *Game) noteBorrowedActivationLocked(key string) {
	if key == "" {
		return
	}
	if g.Activations.Turn == nil {
		g.Activations.Turn = map[string]int{}
	}
	g.Activations.Turn[key]++
}

// BorrowedLabel is a row's label as the borrower has it (CR 201.5b): each
// instance of the lender's name becomes the borrower's. An own or granted
// row is returned unchanged.
func BorrowedLabel(label string, o AbilityOrigin, borrowerName string) string {
	if !o.Borrowed || o.Lent == nil {
		return label
	}
	from := o.Lent.LenderName
	if from == "" || borrowerName == "" || from == borrowerName {
		return label
	}
	return strings.ReplaceAll(label, from, borrowerName)
}

// BorrowInfo is one borrowing static an object is under, for the wire's
// granted_abilities list (ADR 0155 Decision 10): its printed clause and
// its source.
type BorrowInfo struct {
	Label  string
	Source uuid.UUID
}

// BorrowsOf lists each borrowing static that lends `c` at least one row
// right now, once, in layer order.
func BorrowsOf(c Card) []BorrowInfo {
	list := layeredBorrows(&c)
	if len(list) == 0 {
		return nil
	}
	var out []BorrowInfo
	seen := map[string]bool{}
	for _, b := range list {
		if seen[b.Borrow] {
			continue
		}
		seen[b.Borrow] = true
		if label := BorrowLabelOf(b.Borrow); label != "" {
			out = append(out, BorrowInfo{Label: label, Source: b.Source})
		}
	}
	return out
}

// --- the lender the engine does not know (ADR 0155 owner decision 2) ------

// UnautomatedLendersOf lists the lenders of the borrows `c` makes for
// itself whose abilities the engine cannot lend: an object with no catalog
// entry and no CR 305.6 land type, so it lends nothing and the borrower's
// menu names it in a disabled "<lender>: not automated" line instead.
//
// A face-down object is left out (CR 708.2a: it has no abilities to lend),
// and so is a token with no entry, which would be one line per Soldier. A
// borrower that has lost its abilities, or whose designation gate is
// closed, borrows nothing and names nothing. Self-borrowers only: the
// recipient is the static's own source, which is every card of the
// battlefield and graveyard family. The statics are the catalog's, so this
// is read-only and safe under the view's read lock.
func (g *Game) UnautomatedLendersOf(c *Card) []*Card {
	if c == nil || c.faceDownPermanent() || c.HasLostAllAbilities() {
		return nil
	}
	var out []*Card
	for _, ab := range staticAbilitiesOf(c) {
		if ab.Borrows == nil || ab.Borrows.Lenders == nil || ab.AppliesTo == nil ||
			!StaticFunctionsFromZone(ab, ZoneBattlefield) || !ab.AppliesTo(c, g, c) {
			continue
		}
		for _, l := range ab.Borrows.Lenders(g, c) {
			if l == nil || l.InstanceID == c.InstanceID || !lenderUnautomated(l) {
				continue
			}
			out = append(out, l)
		}
	}
	return out
}

// lenderUnautomated reports whether a lender lends nothing because the
// engine has no entry for it.
func lenderUnautomated(l *Card) bool {
	if l.FaceDown || IsTokenPtr(l) {
		return false
	}
	if key := catalogKeyOf(l); key != "" && CatalogLookup != nil && CatalogLookup(key) != nil {
		return false
	}
	// A basic land has no entry and still lends its CR 305.6 mana ability.
	return len(lentRowsOf(l, 0)) == 0
}
