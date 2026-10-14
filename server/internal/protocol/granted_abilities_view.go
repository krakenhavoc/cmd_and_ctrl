package protocol

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_abilities_view.go — ADR 0093 Decision 8: how an ability one
// permanent GAVE another reaches the wire.
//
// Three things ship, all public (the grantor is on the battlefield or
// is a spell that resolved in front of everyone):
//
//   - `ref` on every ActivatedAbilityView and ManaAbilityView row, the
//     row's stable name, sent back beside the index so a grant that
//     appears or vanishes between view and announce is refused rather
//     than fired on the wrong row (Decision 5);
//   - `granted_by: {id, name}` on a granted row, which is what the
//     client reads to open its picker on left-click;
//   - `granted_abilities: [{text, source_id, source_name}]` on the
//     CardView, the bundle's printed text — the only way a granted
//     TRIGGER, which has no row, is visible at all.

// GrantedByView names the object that granted an ability row.
type GrantedByView struct {
	ID string `json:"id"`
	// Name is the grantor's name, or empty when it is no longer on the
	// battlefield to be named.
	Name string `json:"name,omitempty"`

	// knowers is set on a LENDER that sits in a hidden zone (ADR 0155
	// Decision 10): the seats that may see it there, keyed by seat UUID
	// string. Nil for a public lender and for every grantor. FilterViewFor
	// drops the entry for a viewer who is not a knower
	// (applyLenderKnowersFor). Unexported, so never on the wire.
	knowers map[string]bool
}

// GrantedAbilityView is one ability another effect gave a permanent,
// as its granting card prints it.
type GrantedAbilityView struct {
	// Text is the quoted ability — "{T}: Add one mana of any color."
	Text string `json:"text"`
	// SourceID / SourceName name the granting object. Both are absent
	// for a copy's CR 707.9a grant, which is part of the permanent's
	// own copiable values and has no granting object on the board.
	SourceID   string `json:"source_id,omitempty"`
	SourceName string `json:"source_name,omitempty"`
}

// grantedByView builds a row's granted_by, or nil for an own row.
// The name is read off the grantor while it is still findable; a
// grantor that has left (never the case for a static grant, whose
// grantor's presence is the grant) is named by ID alone.
func grantedByView(g *game.Game, o game.AbilityOrigin) *GrantedByView {
	if !o.Granted() && !o.Borrowed {
		return nil
	}
	v := &GrantedByView{}
	if o.GrantedBy != uuid.Nil {
		v.ID = o.GrantedBy.String()
		if g != nil {
			if src, ok := g.LookupCardForEffect(o.GrantedBy); ok {
				v.Name = src.Name
			}
		}
	}
	return v
}

// stampGrantedAbilities is the battlefield pass's half of Decision 8
// that needs the game handle: the grantor's NAME on each granted mana
// row (viewOfManaAbilities runs without a game and stamps the ID), and
// the CardView's granted_abilities list. Runs under the read lock
// ViewOfGame already holds.
func stampGrantedAbilities(g *game.Game, card game.Card, c *CardView) {
	for i := range c.ManaAbilities {
		gb := c.ManaAbilities[i].GrantedBy
		if gb == nil || gb.Name != "" || gb.ID == "" {
			continue
		}
		id, err := uuid.Parse(gb.ID)
		if err != nil {
			continue
		}
		if src, ok := g.LookupCardForEffect(id); ok {
			named := *gb
			named.Name = src.Name
			c.ManaAbilities[i].GrantedBy = &named
		}
	}
	// ADR 0155 Decision 10: a borrowed mana row's lender, marked with
	// who may see it when it sits in a hidden zone.
	for i := range c.ManaAbilities {
		if bf := c.ManaAbilities[i].BorrowedFrom; bf != nil && bf.knowers == nil {
			c.ManaAbilities[i].BorrowedFrom = withLenderKnowers(g, *bf)
		}
	}
	c.GrantedAbilities = viewOfGrantedAbilities(g, card)
	c.UnautomatedLenders = viewOfUnautomatedLenders(g, &card)
}

// viewOfUnautomatedLenders projects game.UnautomatedLendersOf: each lender
// the catalog cannot lend from, by id and printed name, once per lender.
// A lender in a hidden zone carries its knowers, as borrowed_from does,
// and reaches only them.
func viewOfUnautomatedLenders(g *game.Game, card *game.Card) []GrantedByView {
	lenders := g.UnautomatedLendersOf(card)
	if len(lenders) == 0 {
		return nil
	}
	out := make([]GrantedByView, 0, len(lenders))
	seen := make(map[uuid.UUID]bool, len(lenders))
	for _, l := range lenders {
		if seen[l.InstanceID] {
			continue
		}
		seen[l.InstanceID] = true
		out = append(out, *withLenderKnowers(g, GrantedByView{ID: l.InstanceID.String(), Name: l.Name}))
	}
	return out
}

// borrowedFromView builds a borrowed row's borrowed_from: the lender's
// ID and printed name, which the entry carries (ADR 0155 Decision 10).
// Nil for a row that is not lent. With a game handle, a lender in a
// hidden zone carries the seats that may see it; without one
// (viewOfManaAbilitiesFromZone), stampGrantedAbilities asks afterwards.
func borrowedFromView(g *game.Game, o game.AbilityOrigin) *GrantedByView {
	if !o.Borrowed || o.Lent == nil || o.LentBy == uuid.Nil {
		return nil
	}
	v := GrantedByView{ID: o.LentBy.String(), Name: o.Lent.LenderName}
	if g == nil {
		return &v
	}
	return withLenderKnowers(g, v)
}

// withLenderKnowers marks a lender that sits in a zone whose contents are
// hidden (a hand or a library) with the seats that may see it there
// (game.LenderKnowersLocked): a library top revealed to the table
// (Conspicuous Snoop, Skill Borrower) is every seat's, a card in a hand
// its owner's and the seats it was revealed to. FilterViewFor drops the
// lender for every other viewer. A public lender comes back unmarked.
func withLenderKnowers(g *game.Game, v GrantedByView) *GrantedByView {
	lid, err := uuid.Parse(v.ID)
	if err != nil {
		v.knowers = map[string]bool{}
		return &v
	}
	hidden, seats := g.LenderKnowersLocked(lid)
	if !hidden {
		return &v
	}
	v.knowers = make(map[string]bool, len(seats))
	for _, s := range seats {
		v.knowers[s.String()] = true
	}
	return &v
}

// lenderVisibleTo reports whether a viewer may read a lender: every
// viewer for a public one, the admin always (FilterViewFor's empty
// viewer), a spectator only when every seat may (knownToEverySeat's
// rule), and a seat when it is a knower.
func lenderVisibleTo(v *GrantedByView, viewerID string, seatIDs []string) bool {
	if v == nil || v.knowers == nil || viewerID == "" {
		return true
	}
	if viewerID == SpectatorViewerID {
		if len(seatIDs) == 0 {
			return false
		}
		for _, id := range seatIDs {
			if !v.knowers[id] {
				return false
			}
		}
		return true
	}
	return v.knowers[viewerID]
}

// applyLenderKnowersFor is FilterViewFor's per-viewer half of ADR 0155
// Decision 10 over a public zone's cards (the battlefield): a borrowed
// row whose lender this viewer may not see loses its borrowed_from, and
// an unautomated lender this viewer may not see is dropped. The row is
// still there, with the borrower's own label: what the permanent can do
// is public, and which hidden card lent it is not.
//
// It copies before it writes. The rows and the lender lists share their
// backing arrays with the unfiltered view and with every other viewer's
// copy, so writing in place would hand this viewer's answer to the
// next. Nothing is copied for a card with no hidden lender, which is
// every card but a Snoop's or a Skill Borrower's.
func applyLenderKnowersFor(z ZoneView, viewerID string, seatIDs []string) ZoneView {
	for i := range z.Cards {
		c := &z.Cards[i]
		c.ActivatedAbilities = hideLendersInActivated(c.ActivatedAbilities, viewerID, seatIDs)
		c.ManaAbilities = hideLendersInMana(c.ManaAbilities, viewerID, seatIDs)
		if len(c.UnautomatedLenders) > 0 {
			var kept []GrantedByView
			drop := false
			for j := range c.UnautomatedLenders {
				if !lenderVisibleTo(&c.UnautomatedLenders[j], viewerID, seatIDs) {
					drop = true
					break
				}
			}
			if drop {
				for j := range c.UnautomatedLenders {
					if lenderVisibleTo(&c.UnautomatedLenders[j], viewerID, seatIDs) {
						kept = append(kept, c.UnautomatedLenders[j])
					}
				}
				c.UnautomatedLenders = kept
			}
		}
	}
	return z
}

func hideLendersInActivated(rows []ActivatedAbilityView, viewerID string, seatIDs []string) []ActivatedAbilityView {
	var out []ActivatedAbilityView
	for i := range rows {
		if lenderVisibleTo(rows[i].BorrowedFrom, viewerID, seatIDs) {
			continue
		}
		if out == nil {
			out = append([]ActivatedAbilityView(nil), rows...)
		}
		out[i].BorrowedFrom = nil
	}
	if out == nil {
		return rows
	}
	return out
}

func hideLendersInMana(rows []ManaAbilityView, viewerID string, seatIDs []string) []ManaAbilityView {
	var out []ManaAbilityView
	for i := range rows {
		if lenderVisibleTo(rows[i].BorrowedFrom, viewerID, seatIDs) {
			continue
		}
		if out == nil {
			out = append([]ManaAbilityView(nil), rows...)
		}
		out[i].BorrowedFrom = nil
	}
	if out == nil {
		return rows
	}
	return out
}

// viewOfGrantedAbilities projects game.GrantedAbilitiesOf.
func viewOfGrantedAbilities(g *game.Game, card game.Card) []GrantedAbilityView {
	infos := game.GrantedAbilitiesOf(card)
	if len(infos) == 0 {
		return nil
	}
	out := make([]GrantedAbilityView, 0, len(infos))
	for _, info := range infos {
		v := GrantedAbilityView{Text: info.Text}
		if info.Source != uuid.Nil {
			v.SourceID = info.Source.String()
			if src, ok := g.LookupCardForEffect(info.Source); ok {
				v.SourceName = src.Name
			}
		}
		out = append(out, v)
	}
	return out
}
