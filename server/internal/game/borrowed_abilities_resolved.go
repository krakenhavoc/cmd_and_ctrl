package game

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// borrowed_abilities_resolved.go — ADR 0155 Decision 9: a borrow made by
// a RESOLVING ability, as data. The static borrow is borrowed_abilities.go.

// ModBorrowAbilities gives each affected object a fixed set of lent rows
// for the record's duration: "This creature gains all activated
// abilities of target creature until end of turn" (Quicksilver
// Elemental), "each Horror you control gains all activated abilities of
// target artifact an opponent controls until end of turn" (Grell
// Philosopher), Havengul Lich's "gains all activated abilities of that
// card until end of turn".
//
// A resolving ability fixes both sets once, as it resolves (CR 608.2h,
// CR 611.2c): the affected set is the record's, and the lent set is the
// rows the lender had then, captured as data in Mod.Borrows
// (LentRowsForEffect). Quicksilver Elemental keeps what it gained after
// its target leaves. Each row names the catalog row it came from, so on
// the recipient it is the static borrow's BorrowedAbility: the same
// `borrow:` refs, the same stack stamp, the same CR 201.5b label.
//
// Layer 6, in the record's TIMESTAMP slot rather than the borrow step: a
// fixed set depends on nothing, so CR 613.7's timestamp order is all
// there is. A removal on the recipient takes the rows when it is later
// and leaves them when it is earlier, as it does a grant.
//
// The kind string and Mod.Borrows are on-disk identities, additive within
// the schema version (ADR 0093 PR 4's precedent). A restore point naming
// a row this binary's catalog does not have is refused with
// ErrUnknownEffectKey, as an unknown grant bundle is.
const ModBorrowAbilities ModKind = "borrowAbilities" // layer 6

// LentRow is one row a resolved borrow lends: BorrowedAbility's row
// identity, plus the lender it was read off, for the wire. Plain data,
// carried in Mod.Borrows.
type LentRow struct {
	// Slot is AbilitySlotActivated or BorrowSlotMana.
	Slot string `json:"slot"`
	// Kind, Key, Index and Color are BorrowedAbility's.
	Kind  LentKind `json:"kind,omitempty"`
	Key   string   `json:"key,omitempty"`
	Index int      `json:"index,omitempty"`
	Color string   `json:"color,omitempty"`
	// Lender is the object the row was read off and LenderName its
	// printed name then: the wire's borrowed_from and CR 201.5b's
	// rewrite, after the lender has gone.
	Lender     uuid.UUID `json:"lender"`
	LenderName string    `json:"lenderName,omitempty"`
}

// BorrowAbilitiesMod is "gains all activated abilities of <lender>" from
// a resolving ability (Decision 9): the rows LentRowsForEffect captured.
// `spendFrom` is the resolved effect's own "you may spend <colour> mana
// as though it were mana of any color to activate those abilities"
// (Grell Philosopher's "U"), or empty. Registration refuses a mod with
// no row.
func BorrowAbilitiesMod(spendFrom string, rows ...LentRow) Mod {
	return Mod{Kind: ModBorrowAbilities, Borrows: append([]LentRow(nil), rows...), SpendFrom: spendFrom}
}

// LentRowsForEffect lists the rows `lender` lends right now, as data for
// BorrowAbilitiesMod: on the battlefield its finished layer-6 rows (a
// grant on it, or a row it borrowed itself, included), anywhere else its
// printed ones (Decision 3). Nil for a face-down or uncatalogued lender.
// Caller must hold g.mu (a resolving effect does).
func LentRowsForEffect(lender *Card) []LentRow {
	return lentRowsAsData(lentRowsOf(lender, 0))
}

// PrintedLentRowsForEffect is LentRowsForEffect read off the card as
// printed, whatever zone it is in now: Havengul Lich's "that card", the
// spell as it was cast, not the permanent it may since have become.
func PrintedLentRowsForEffect(lender Card) []LentRow {
	lender.effective = nil
	return lentRowsAsData(lentRowsOf(&lender, 0))
}

// UnautomatedBorrowMod is the borrowAbilities mod for a lender the
// catalog has no entry for (ADR 0155 owner decision 2): it lends nothing,
// and names the lender so the recipient's menu can say "<lender>: not
// automated" for as long as the effect lasts. Nil when the lender does
// lend, or lends nothing for another reason (a face-down object, a
// token, a card with no activated ability).
func UnautomatedBorrowMod(lender *Card) *Mod {
	if lender == nil || !lenderUnautomated(lender) {
		return nil
	}
	return &Mod{Kind: ModBorrowAbilities, Unautomated: []LentRow{{Lender: lender.InstanceID, LenderName: lender.Name}}}
}

func lentRowsAsData(rows []BorrowedAbility) []LentRow {
	if len(rows) == 0 {
		return nil
	}
	out := make([]LentRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, LentRow{
			Slot: r.Slot, Kind: r.Kind, Key: r.Key, Index: r.Index, Color: r.Color,
			Lender: r.Lender, LenderName: r.LenderName,
		})
	}
	return out
}

// borrowModApply is the layer adapter's Apply for a borrowAbilities mod:
// the captured rows, written as BorrowedAbility entries on the recipient
// with the record's source as the lending effect's source and the
// record's label as the borrow's clause. No resolved borrow prints "only
// once each turn", so no row is.
func borrowModApply(m Mod, label string) func(*Characteristic, *Card, *Game, *Card) {
	rows := m.Borrows
	spendFrom := m.SpendFrom
	return func(ch *Characteristic, _ *Card, _ *Game, src *Card) {
		var from, you uuid.UUID
		if src != nil {
			from, you = src.InstanceID, src.Controller
		}
		id := borrowIdentity(from, label)
		for _, r := range rows {
			ch.BorrowedAbilities = append(ch.BorrowedAbilities, BorrowedAbility{
				Slot: r.Slot, Kind: r.Kind, Key: r.Key, Index: r.Index, Color: r.Color,
				Lender: r.Lender, LenderName: r.LenderName,
				Source: from, Activator: you, Borrow: id, SpendFrom: spendFrom,
			})
		}
	}
}

// borrowModProblem is the registration check for a borrowAbilities mod,
// and for the borrow fields on any other kind. Empty when the mod is
// well formed.
func borrowModProblem(m Mod) string {
	if m.Kind != ModBorrowAbilities {
		if len(m.Borrows) > 0 || len(m.Unautomated) > 0 || m.SpendFrom != "" {
			return fmt.Sprintf("mod %q carries lent rows, which only %q reads", m.Kind, ModBorrowAbilities)
		}
		return ""
	}
	if len(m.Borrows) == 0 && len(m.Unautomated) == 0 {
		return "lends no ability row and names no lender"
	}
	for _, u := range m.Unautomated {
		if u.Lender == uuid.Nil {
			return "names an unautomated lender with no ID"
		}
	}
	if m.SpendFrom != "" && (len(m.SpendFrom) != 1 || !strings.Contains("WUBRG", m.SpendFrom)) {
		return fmt.Sprintf("spends %q mana as though it were any color — want one of W, U, B, R, G", m.SpendFrom)
	}
	for _, r := range m.Borrows {
		if r.Slot != AbilitySlotActivated && r.Slot != BorrowSlotMana {
			return fmt.Sprintf("lends a row in slot %q", r.Slot)
		}
		switch r.Kind {
		case LentCatalog:
			if r.Key == "" || r.Index < 0 {
				return "lends a catalog row with no key"
			}
		case LentLand:
			if _, ok := intrinsicLandRowFor(r.Color); !ok || r.Slot != BorrowSlotMana {
				return fmt.Sprintf("lends a land mana row of colour %q", r.Color)
			}
		default:
			return fmt.Sprintf("lends a row of unknown kind %d", r.Kind)
		}
	}
	return ""
}

// unknownLentRows lists the rows of a borrowAbilities mod this binary
// cannot read: a malformed mod, or a catalog row this binary's catalog
// does not have — a restore point a newer build wrote, refused with
// ErrUnknownEffectKey (Decision 9).
func unknownLentRows(m Mod) []string {
	if m.Kind != ModBorrowAbilities {
		return nil
	}
	if problem := borrowModProblem(m); problem != "" {
		return []string{"borrowAbilities mod " + problem}
	}
	var out []string
	for _, r := range m.Borrows {
		if r.Kind != LentCatalog {
			continue
		}
		n := 0
		if catalogDef(r.Key) != nil {
			switch {
			case r.Slot == BorrowSlotMana && CatalogManaAbilities != nil:
				n = len(CatalogManaAbilities(r.Key))
			case r.Slot == AbilitySlotActivated && CatalogActivatedAbilities != nil:
				n = len(CatalogActivatedAbilities(r.Key))
			}
		}
		if r.Index >= n {
			out = append(out, fmt.Sprintf("lent %s row %q:%d", r.Slot, r.Key, r.Index))
		}
	}
	return out
}
