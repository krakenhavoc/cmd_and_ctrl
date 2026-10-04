package protocol

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// legal_actions.go — ADR 0105 §1, sub-PR 1 (#1789).
//
// LegalActionsView is a per-card digest of the viewer's own legal-move
// enumeration. It exists so the client can highlight every action the
// seat may take — a castable card, a permanent's live ability row, a
// creature that may attack — from the SAME answer the server accepts
// clicks against.
//
// It is built from the UNCAPPED list, before capLegalMoves degrades an
// over-long one to one move per (source, kind, targets_stack). So a
// board past legalMovesWireCap still says which ability rows are live,
// which faces are castable and from which zones, which the capped
// legal_moves cannot.
//
// It is not a second answer to "is this legal". digestLegalMoves reads
// a []legal.Move and nothing else: no game read, no lock.

// LegalActionsView is GameView.legal_actions: the digest of the
// viewer's own uncapped move list.
type LegalActionsView struct {
	// Pass is true when the list holds a pass_priority move — the
	// question the client's hasPassMove answers by scanning today.
	Pass bool `json:"pass,omitempty"`
	// Sources maps a card instance ID to what that card may do.
	Sources map[string]*LegalSourceView `json:"sources,omitempty"`
	// order is the order the sources were first enumerated in, which
	// is the order MarshalJSON writes them in.
	order []string
}

// MarshalJSON writes the digest with its sources in enumeration order
// rather than encoding/json's sorted-key order.
//
// The keys are random instance UUIDs, so sorted order differs from one
// game to the next even when the boards are identical. A frame could
// then not be compared byte for byte once its IDs are replaced with
// placeholders, and that comparison is what keeps the committed
// agreement fixtures in internal/legal/testdata honest. Enumeration
// order is fixed by the board. A view that was decoded rather than
// built has no recorded order and falls back to sorted keys.
func (v LegalActionsView) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	if v.Pass {
		buf.WriteString(`"pass":true`)
	}
	if len(v.Sources) > 0 {
		if v.Pass {
			buf.WriteByte(',')
		}
		buf.WriteString(`"sources":{`)
		for i, id := range v.sourceOrder() {
			if i > 0 {
				buf.WriteByte(',')
			}
			key, err := json.Marshal(id)
			if err != nil {
				return nil, err
			}
			val, err := json.Marshal(v.Sources[id])
			if err != nil {
				return nil, err
			}
			buf.Write(key)
			buf.WriteByte(':')
			buf.Write(val)
		}
		buf.WriteByte('}')
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// sourceOrder is every key of Sources exactly once: the recorded
// enumeration order first, then any key it does not name, sorted.
func (v LegalActionsView) sourceOrder() []string {
	out := make([]string, 0, len(v.Sources))
	seen := make(map[string]bool, len(v.Sources))
	for _, id := range v.order {
		if _, ok := v.Sources[id]; ok && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	var rest []string
	for id := range v.Sources {
		if !seen[id] {
			rest = append(rest, id)
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}

// LegalSourceView is one card's entry in the digest. Every slice is in
// enumeration order with duplicates removed.
type LegalSourceView struct {
	// Kinds are the distinct legal.Kind values of the card's moves.
	// Never "pass", "choice" or "mulligan": those are not digested.
	Kinds []legal.Kind `json:"kinds"`
	// Moves is how many moves in the uncapped list involve this card:
	// its own moves (Source), plus the grouped block declarations it
	// is one of the blockers in.
	Moves int `json:"moves"`
	// Abilities are the ADR 0093 row refs of the card's live
	// activate_ability moves — the `ref` on an activated_abilities or
	// zone_abilities row.
	Abilities []string `json:"abilities,omitempty"`
	// ManaAbilities are the refs of its live activate_mana_ability
	// moves — the `ref` on a mana_abilities or zone_mana_abilities row.
	ManaAbilities []string `json:"mana_abilities,omitempty"`
	// SpecialActions are the kinds of its live special_action moves
	// ("foretell", "suspend", "turn_face_up", "plot").
	SpecialActions []string `json:"special_actions,omitempty"`
	// Zones are the from_zone values of its cast and land moves.
	Zones []string `json:"zones,omitempty"`
	// Faces are the printed faces its cast moves cast.
	Faces []int `json:"faces,omitempty"`
	// AttackTargets are the players, planeswalkers and battles this
	// creature may attack.
	AttackTargets []string `json:"attack_targets,omitempty"`
	// Blocks are the attackers this creature may block, alone or as
	// part of a grouped declaration.
	Blocks []string `json:"blocks,omitempty"`
	// CastIdleHint is set when EVERY cast move for the card carries a
	// legal.Move.IdleHint (#1918): the card is castable, but no cast it
	// has would do anything right now — an overloaded Counterflux with
	// no spell to counter. It is the first such move's hint, and the
	// client draws a muted ring with it as the tooltip. One cast that
	// would do something (the targeted half, with a target) and it is
	// absent, so the card highlights as an ordinary castable card.
	CastIdleHint string `json:"cast_idle_hint,omitempty"`
	// Truncated is the enumerator's own cut report for this card (ADR
	// 0122 §6.2): each count cap that left some of its moves out, and
	// how many. Absent when nothing was cut, which is nearly every
	// card. Moves counts what survived the caps; this says the card has
	// more. A legal_moves_request naming the card returns them, up to
	// legal.ExpandCeiling.
	Truncated []LegalCutView `json:"truncated,omitempty"`
}

// LegalCutView is one entry of the cut report (ADR 0122 §6.2), on the
// wire in the full-list reply's `truncated` and, per card, in the
// digest. See legal.Cut for what the numbers mean.
type LegalCutView struct {
	// Source is the card whose expansion was cut. Omitted for a prompt
	// no card raised, and inside the digest, where the key is the card.
	Source string `json:"source,omitempty"`
	// Choice is the pending choice being answered — its ID, or
	// "cleanup_discard" — when the cut was in one.
	Choice string `json:"choice,omitempty"`
	// Cap names the cap: per_source, max_x, variable_counts,
	// subset_scan, creature_types, card_names, cost_payments, repeats,
	// or ceiling (an expanded request's 512).
	Cap string `json:"cap"`
	// Omitted is how many candidate moves the cap left out; a lower
	// bound when AtLeast is set.
	Omitted int  `json:"omitted"`
	AtLeast bool `json:"at_least,omitempty"`
}

// LegalCutViews converts the enumerator's cut report to its wire form.
func LegalCutViews(cuts []legal.Cut) []LegalCutView {
	if len(cuts) == 0 {
		return nil
	}
	out := make([]LegalCutView, 0, len(cuts))
	for _, c := range cuts {
		out = append(out, legalCutView(c, true))
	}
	return out
}

func legalCutView(c legal.Cut, withSource bool) LegalCutView {
	v := LegalCutView{Choice: c.Choice, Cap: string(c.Cap), Omitted: c.Omitted, AtLeast: c.AtLeast}
	if withSource && c.Source != uuid.Nil {
		v.Source = c.Source.String()
	}
	return v
}

// fileDigestCuts files each card's cuts under that card's digest entry
// (ADR 0122 §6.2). Only a card the digest already lists gets them: an
// entry is a highlight, so a cut never invents one, and a prompt's cut
// (whose answers the digest does not fold) rides the full-list reply
// alone.
func fileDigestCuts(d *LegalActionsView, cuts []legal.Cut) {
	if d == nil {
		return
	}
	for _, c := range cuts {
		if c.Source == uuid.Nil || c.Choice != "" {
			continue
		}
		e := d.Sources[c.Source.String()]
		if e == nil {
			continue
		}
		e.Truncated = append(e.Truncated, legalCutView(c, false))
	}
}

// idleTally is digestLegalMoves' per-card count behind CastIdleHint.
type idleTally struct {
	casts, idle int
	hint        string
}

// The move params the digest reads. Each is a subset of the params
// struct internal/legal marshals for that move type; the json names
// are the wire contract docs/protocol.md documents for each action.
type (
	digestCastParams struct {
		FromZone string `json:"from_zone"`
		Face     int    `json:"face"`
	}
	digestAbilityParams struct {
		Ref string `json:"ref"`
	}
	digestSpecialParams struct {
		Kind string `json:"kind"`
	}
	digestAttackParams struct {
		Target string `json:"target"`
	}
	digestBlockParams struct {
		Blocker  string `json:"blocker"`
		Attacker string `json:"attacker"`
	}
	digestBlocksParams struct {
		Blocks []digestBlockParams `json:"blocks"`
	}
)

// digestLegalMoves folds one seat's uncapped move list into its
// LegalActionsView. Nil when the list gives the client nothing to
// highlight — no moves at all, or only choice, mulligan and opening-roll
// answers, whose surfaces read pending_choices, the mulligan window and
// opening_roll — so a quiet frame costs 0 bytes.
func digestLegalMoves(moves []legal.Move) *LegalActionsView {
	var out *LegalActionsView
	// #1918: cast moves per card, and how many of them are idle.
	var idle map[*LegalSourceView]*idleTally
	view := func() *LegalActionsView {
		if out == nil {
			out = &LegalActionsView{}
		}
		return out
	}
	entry := func(id string) *LegalSourceView {
		v := view()
		if v.Sources == nil {
			v.Sources = make(map[string]*LegalSourceView)
		}
		e := v.Sources[id]
		if e == nil {
			e = &LegalSourceView{}
			v.Sources[id] = e
			v.order = append(v.order, id)
		}
		return e
	}
	for i := range moves {
		m := &moves[i]
		switch m.Kind {
		case legal.KindPass:
			view().Pass = true
			continue
		case legal.KindChoice, legal.KindMulligan, legal.KindOpeningRoll:
			continue
		}
		if m.Source == uuid.Nil {
			continue
		}
		e := entry(m.Source.String())
		e.Moves++
		e.Kinds = appendUnique(e.Kinds, m.Kind)
		switch m.Kind {
		case legal.KindCast, legal.KindLand:
			var p digestCastParams
			if json.Unmarshal(m.Params, &p) == nil {
				e.Zones = appendUniqueNonEmpty(e.Zones, p.FromZone)
				if m.Kind == legal.KindCast {
					e.Faces = appendUnique(e.Faces, p.Face)
				}
			}
			if m.Kind == legal.KindCast {
				if idle == nil {
					idle = make(map[*LegalSourceView]*idleTally)
				}
				tl := idle[e]
				if tl == nil {
					tl = &idleTally{}
					idle[e] = tl
				}
				tl.casts++
				if m.IdleHint != "" {
					tl.idle++
					if tl.hint == "" {
						tl.hint = m.IdleHint
					}
				}
			}
		case legal.KindActivate:
			var p digestAbilityParams
			if json.Unmarshal(m.Params, &p) == nil {
				e.Abilities = appendUniqueNonEmpty(e.Abilities, p.Ref)
			}
		case legal.KindMana:
			var p digestAbilityParams
			if json.Unmarshal(m.Params, &p) == nil {
				e.ManaAbilities = appendUniqueNonEmpty(e.ManaAbilities, p.Ref)
			}
		case legal.KindSpecialAction:
			var p digestSpecialParams
			if json.Unmarshal(m.Params, &p) == nil {
				e.SpecialActions = appendUniqueNonEmpty(e.SpecialActions, p.Kind)
			}
		case legal.KindAttack:
			var p digestAttackParams
			if json.Unmarshal(m.Params, &p) == nil {
				e.AttackTargets = appendUniqueNonEmpty(e.AttackTargets, p.Target)
			}
		case legal.KindBlock:
			digestBlockMove(m, e, entry)
		}
	}
	for e, tl := range idle {
		if tl.idle == tl.casts {
			e.CastIdleHint = tl.hint
		}
	}
	return out
}

// digestBlockMove records a block move's (blocker, attacker) pairs. A
// declare_blocker names one pair. A declare_blockers names a group —
// the two creatures a menace attacker takes (#750) — and every
// creature in the group is a block candidate, not only the one the
// move's Source names, so each blocker gets the pair it takes part in.
func digestBlockMove(m *legal.Move, src *LegalSourceView, entry func(string) *LegalSourceView) {
	if m.Type == legal.TypeDeclareBlockers {
		var p digestBlocksParams
		if json.Unmarshal(m.Params, &p) != nil {
			return
		}
		source := m.Source.String()
		counted := map[string]bool{source: true}
		for _, b := range p.Blocks {
			if b.Blocker == "" {
				continue
			}
			e := src
			if b.Blocker != source {
				e = entry(b.Blocker)
				if !counted[b.Blocker] {
					counted[b.Blocker] = true
					e.Moves++
					e.Kinds = appendUnique(e.Kinds, legal.KindBlock)
				}
			}
			e.Blocks = appendUniqueNonEmpty(e.Blocks, b.Attacker)
		}
		return
	}
	var p digestBlockParams
	if json.Unmarshal(m.Params, &p) == nil {
		src.Blocks = appendUniqueNonEmpty(src.Blocks, p.Attacker)
	}
}

func appendUnique[T comparable](s []T, v T) []T {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}

func appendUniqueNonEmpty(s []string, v string) []string {
	if v == "" {
		return s
	}
	return appendUnique(s, v)
}

// legalActionsFor picks the viewer's own digest out of the per-seat
// map, under exactly legalMovesFor's rule: the admin (""), a spectator
// (SpectatorViewerID) and any viewer who is not a seat get nothing.
func legalActionsFor(bySeat map[string]*LegalActionsView, viewerID string) *LegalActionsView {
	if viewerID == "" || viewerID == SpectatorViewerID || len(bySeat) == 0 {
		return nil
	}
	return bySeat[viewerID]
}
