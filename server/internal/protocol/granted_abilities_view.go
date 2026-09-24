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
	if !o.Granted() {
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
	c.GrantedAbilities = viewOfGrantedAbilities(g, card)
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
