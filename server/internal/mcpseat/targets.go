package mcpseat

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// targets.go — #2276: say who a target is, and let the agent pick targets
// from the board per clause instead of from the enumerator's capped
// combinations.
//
// Nothing here needs a new frame. A move's Params already carry its
// targets ({kind, id, slot, mode}), the view already carries each clause's
// legal set and bounds (CardView.legal_targets / clauses, an ability's,
// a mode option's, a pick_target prompt's), and the server already
// validates the targets an `action` names with the same check a person's
// targeting goes through (CR 601.2c). The seat reads those, shows them,
// and swaps the agent's picks into the chosen move's params.

// wireTarget is one target slot as a move's params carry it.
type wireTarget struct {
	Kind string `json:"kind"`
	ID   string `json:"id,omitempty"`
	Slot int    `json:"slot,omitempty"`
	Mode int    `json:"mode,omitempty"`
}

// TargetPick is act's `targets` entry: the picks for one target clause.
type TargetPick struct {
	Slot int      `json:"slot,omitempty" jsonschema:"the target clause, from 0 (default 0), as legal_moves(targets_for) numbers them"`
	Mode int      `json:"mode,omitempty" jsonschema:"for a modal move that chose the same mode twice or more: which occurrence, from 0 (default 0)"`
	IDs  []string `json:"ids" jsonschema:"the ids of the players or cards picked for this clause"`
}

// moveTargets is the targets a move's params already name.
func moveTargets(m legal.Move) []wireTarget {
	if len(m.Params) == 0 {
		return nil
	}
	var p struct {
		Targets []wireTarget `json:"targets"`
		Target  *wireTarget  `json:"target"`
	}
	if json.Unmarshal(m.Params, &p) != nil {
		return nil
	}
	if len(p.Targets) > 0 {
		return p.Targets
	}
	if p.Target != nil {
		return []wireTarget{*p.Target}
	}
	return nil
}

// playerRef says who a player is relative to this seat: "you (seat N)" or
// "opponent «name» (seat N)". It goes by the id, never by display name,
// since two seats can share a name.
func playerRef(v *protocol.GameView, me, id string) (string, bool) {
	for i := range v.Seats {
		s := &v.Seats[i]
		if s.ID != id {
			continue
		}
		if id == me {
			return fmt.Sprintf("you (seat %d)", s.Seat), true
		}
		name := s.DisplayName
		if name == "" {
			name = s.Name
		}
		return fmt.Sprintf("opponent %s (seat %d)", untrusted(name, maxNameLen), s.Seat), true
	}
	return "", false
}

// controllerOf is "you" or the opponent's seat, for a card's line.
func controllerOf(v *protocol.GameView, me string, c *protocol.CardView) string {
	if c.Controller == "" {
		return ""
	}
	if c.Controller == me {
		return "yours"
	}
	if r, ok := playerRef(v, me, c.Controller); ok {
		return "controlled by " + r
	}
	return ""
}

// targetRef names one target for the model: a player by who they are, a
// card by name, id, zone and controller.
func targetRef(v *protocol.GameView, me string, kind, id string) string {
	if kind == "self" {
		return "you"
	}
	if kind == "player" || (kind == "" && isSeat(v, id)) {
		if r, ok := playerRef(v, me, id); ok {
			return r
		}
		return "player " + id
	}
	c, zone := findCard(v, id)
	if c == nil {
		return "card [" + id + "]"
	}
	s := boardtext.CardName(c) + " [" + id + "]"
	var bits []string
	if zone != "" {
		bits = append(bits, zone)
	}
	if ctl := controllerOf(v, me, c); ctl != "" {
		bits = append(bits, ctl)
	}
	if len(bits) > 0 {
		s += " (" + strings.Join(bits, ", ") + ")"
	}
	return s
}

func isSeat(v *protocol.GameView, id string) bool {
	for i := range v.Seats {
		if v.Seats[i].ID == id {
			return true
		}
	}
	return false
}

// targetNote is the move line's "targets: …" note: who or what each slot
// names, so "targeting «Agent»" is never the only word on it.
func targetNote(v *protocol.GameView, me string, m legal.Move) string {
	ts := moveTargets(m)
	if len(ts) == 0 || v == nil {
		return ""
	}
	parts := make([]string, 0, len(ts))
	for _, t := range ts {
		parts = append(parts, targetRef(v, me, t.Kind, t.ID))
	}
	return "targets: " + strings.Join(parts, "; ")
}

// clauseRef is one target clause of a move, with where it sits.
type clauseRef struct {
	Mode, Slot int
	protocol.LegalTargetsView
}

func clausesOf(lt *protocol.LegalTargetsView, cl []protocol.LegalTargetsView, mode int) []clauseRef {
	var out []clauseRef
	switch {
	case len(cl) > 0:
		for i, c := range cl {
			out = append(out, clauseRef{Mode: mode, Slot: i, LegalTargetsView: c})
		}
	case lt != nil:
		out = append(out, clauseRef{Mode: mode, LegalTargetsView: *lt})
	}
	return out
}

func modeClauses(ms *protocol.ModeSpecView, chosen []int) []clauseRef {
	if ms == nil {
		return nil
	}
	var out []clauseRef
	for occ, idx := range chosen {
		if idx < 0 || idx >= len(ms.Options) {
			continue
		}
		o := ms.Options[idx]
		out = append(out, clausesOf(o.LegalTargets, o.Clauses, occ)...)
	}
	return out
}

// moveClauses finds the target clauses the view states for a move, nil
// when the view states none for it (a trigger's prompt aside, the cast
// and the activation both ride their card's own surface).
func moveClauses(v *protocol.GameView, m legal.Move) []clauseRef {
	if len(m.Params) == 0 {
		return nil
	}
	var p struct {
		InstanceID   string `json:"instance_id"`
		AltCost      string `json:"alternative_cost"`
		Modes        []int  `json:"modes"`
		SourceCardID string `json:"source_card_id"`
		AbilityIndex int    `json:"ability_index"`
		ChoiceID     string `json:"choice_id"`
	}
	if json.Unmarshal(m.Params, &p) != nil {
		return nil
	}
	switch m.Type {
	case legal.TypeCastSpell:
		c, _ := findCard(v, p.InstanceID)
		if c == nil {
			return nil
		}
		if p.AltCost != "" {
			for _, a := range c.AlternativeCosts {
				if a.Key == p.AltCost && (a.LegalTargets != nil) {
					return clausesOf(a.LegalTargets, nil, 0)
				}
			}
		}
		if len(p.Modes) > 0 {
			if cl := modeClauses(c.Modes, p.Modes); len(cl) > 0 {
				return cl
			}
		}
		return clausesOf(c.LegalTargets, c.Clauses, 0)
	case legal.TypeActivateAbility:
		c, _ := findCard(v, p.SourceCardID)
		if c == nil {
			return nil
		}
		for _, list := range [][]protocol.ActivatedAbilityView{c.ActivatedAbilities, c.ZoneAbilities} {
			for _, a := range list {
				if a.Index != p.AbilityIndex {
					continue
				}
				if len(p.Modes) > 0 {
					if cl := modeClauses(a.Modes, p.Modes); len(cl) > 0 {
						return cl
					}
				}
				return clausesOf(a.LegalTargets, a.Clauses, 0)
			}
		}
	case legal.TypeResolveChoice:
		for i := range v.PendingChoices {
			if pc := &v.PendingChoices[i]; pc.ID == p.ChoiceID && pc.PickTarget != nil {
				return clausesOf(pc.PickTarget, nil, 0)
			}
		}
	}
	return nil
}

// bounds says how many a clause takes, as the model reads it.
func bounds(c protocol.LegalTargetsView) string {
	switch {
	case c.CountFromX && c.UpToX:
		return "up to X (your announced X)"
	case c.CountFromX:
		return "X (your announced X)"
	case c.Max == 0 && c.Min == 0:
		return "any number"
	case c.Max == 0:
		return strconv.Itoa(c.Min) + " or more"
	case c.Min == c.Max:
		return strconv.Itoa(c.Min)
	case c.Min == 0:
		return "up to " + strconv.Itoa(c.Max)
	}
	return fmt.Sprintf("%d to %d", c.Min, c.Max)
}

// targetsText is legal_moves(targets_for: N): the move's target clauses,
// each with its bounds and every candidate on the board, so the agent
// picks the ones it wants and sends them as act's `targets`.
func targetsText(v *protocol.GameView, w *window, n int) (string, error) {
	if n < 0 || n >= len(w.moves) {
		return "", fmt.Errorf("move %d is not in this window's list (0 to %d)", n, len(w.moves)-1)
	}
	m := w.moves[n]
	me := w.me
	nw := newNameWrapper(v)
	var b strings.Builder
	fmt.Fprintf(&b, "TARGETS FOR MOVE %d: %s\n", n, nw.apply(m.Label))
	clauses := moveClauses(v, m)
	source := "the server's legal set for this move"
	if len(clauses) == 0 {
		clauses = unionClauses(v, w, m)
		source = "assembled from this window's move list, which the enumerator may have capped, so a legal target can be missing"
	}
	if len(clauses) == 0 {
		b.WriteString("This move has no target clause the seat can offer picks for. Answer it as listed.\n")
		return b.String(), nil
	}
	b.WriteString("Candidates: " + source + ".\n")
	b.WriteString("Answer with act(window: " + strconv.Quote(w.token) + ", move: " + strconv.Itoa(n) + ", targets: [{slot, ids: [...]}, ...]).\n")
	for _, c := range clauses {
		head := fmt.Sprintf("slot %d", c.Slot)
		if c.Mode > 0 {
			head = fmt.Sprintf("mode occurrence %d, slot %d", c.Mode, c.Slot)
		}
		label := c.Label
		if label == "" {
			label = "target"
		}
		fmt.Fprintf(&b, "%s: %s — pick %s\n", head, nw.apply(label), bounds(c.LegalTargetsView))
		if c.Distinct {
			b.WriteString("  must differ from the earlier clauses' picks\n")
		}
		if c.Different != nil || c.Same != nil {
			b.WriteString("  the picks must also satisfy the clause's set rule; the server checks it\n")
		}
		for _, id := range c.Players {
			fmt.Fprintf(&b, "  player %s — id %s\n", targetRef(v, me, "player", id), id)
		}
		for _, id := range c.Cards {
			fmt.Fprintf(&b, "  %s\n", targetRef(v, me, "card", id))
		}
		if len(c.Players)+len(c.Cards) == 0 {
			b.WriteString("  no legal candidate right now\n")
		}
	}
	return b.String(), nil
}

// unionClauses builds each slot's candidates from the targets of every
// move in the window that is the same move but for its targets. It is the
// fallback for a move whose clause the view does not state.
func unionClauses(v *protocol.GameView, w *window, m legal.Move) []clauseRef {
	if len(moveTargets(m)) == 0 {
		return nil
	}
	key := sansTargets(m)
	type slotKey struct{ mode, slot int }
	cands := map[slotKey]map[string]string{}
	var order []slotKey
	for _, o := range w.moves {
		if o.Type != m.Type || o.Source != m.Source || sansTargets(o) != key {
			continue
		}
		for _, t := range moveTargets(o) {
			k := slotKey{t.Mode, t.Slot}
			if cands[k] == nil {
				cands[k] = map[string]string{}
				order = append(order, k)
			}
			cands[k][t.ID] = t.Kind
		}
	}
	sort.Slice(order, func(i, j int) bool {
		if order[i].mode != order[j].mode {
			return order[i].mode < order[j].mode
		}
		return order[i].slot < order[j].slot
	})
	var out []clauseRef
	for _, k := range order {
		c := clauseRef{Mode: k.mode, Slot: k.slot}
		ids := make([]string, 0, len(cands[k]))
		for id := range cands[k] {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			if cands[k][id] == "player" || isSeat(v, id) {
				c.Players = append(c.Players, id)
			} else {
				c.Cards = append(c.Cards, id)
			}
		}
		out = append(out, c)
	}
	return out
}

// sansTargets is a move's params with the target-bearing keys removed, so
// two moves that differ only in targets compare equal.
func sansTargets(m legal.Move) string {
	var p map[string]any
	if json.Unmarshal(m.Params, &p) != nil {
		return string(m.Params)
	}
	delete(p, "targets")
	delete(p, "target")
	delete(p, "distribution")
	raw, _ := json.Marshal(p)
	return string(raw)
}

// applyTargets swaps the agent's picks into a move's params. The move is
// the template (its card, costs and modes stay), the picks replace its
// targets, and the server checks them as it checks any targeting. A pick
// outside the view's stated legal set, or a count outside its bounds, is
// refused here with the reason, so the model is not left to read a bare
// server code.
func applyTargets(v *protocol.GameView, m legal.Move, params json.RawMessage, picks []TargetPick) (json.RawMessage, error) {
	template := moveTargets(m)
	clauses := moveClauses(v, m)
	if len(template) == 0 && len(clauses) == 0 {
		return nil, fmt.Errorf("this move has no target clause: send it without targets")
	}
	p := map[string]any{}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
	}
	if _, ok := p["distribution"]; ok {
		return nil, fmt.Errorf("this move divides an amount among its targets: choose one of the listed moves")
	}
	type slotKey struct{ mode, slot int }
	byClause := map[slotKey]protocol.LegalTargetsView{}
	for _, c := range clauses {
		byClause[slotKey{c.Mode, c.Slot}] = c.LegalTargetsView
	}
	seen := map[slotKey][]string{}
	var out []wireTarget
	taken := map[string]bool{}
	for _, pk := range picks {
		k := slotKey{pk.Mode, pk.Slot}
		if len(clauses) > 0 {
			if _, ok := byClause[k]; !ok {
				return nil, fmt.Errorf("this move has no target clause at mode %d slot %d (legal_moves(targets_for: <move>) lists them)", pk.Mode, pk.Slot)
			}
		}
		for _, id := range pk.IDs {
			id = strings.TrimSpace(id)
			if id == "" {
				return nil, fmt.Errorf("an empty target id")
			}
			if len(clauses) > 0 && !inClause(byClause[k], id) {
				return nil, fmt.Errorf("%s is not a legal target for slot %d of this move (legal_moves(targets_for: <move>) lists them)", id, pk.Slot)
			}
			if taken[fmt.Sprintf("%d/%d/%s", k.mode, k.slot, id)] {
				return nil, fmt.Errorf("%s is named twice for slot %d", id, pk.Slot)
			}
			taken[fmt.Sprintf("%d/%d/%s", k.mode, k.slot, id)] = true
			kind := "card"
			if isSeat(v, id) {
				kind = "player"
			}
			out = append(out, wireTarget{Kind: kind, ID: id, Slot: pk.Slot, Mode: pk.Mode})
			seen[k] = append(seen[k], id)
		}
	}
	for k, c := range byClause {
		if c.CountFromX {
			continue
		}
		n := len(seen[k])
		if n < c.Min || (c.Max > 0 && n > c.Max) {
			return nil, fmt.Errorf("slot %d takes %s target(s); you sent %d", k.slot, bounds(c), n)
		}
	}
	if len(out) == 0 && len(template) > 0 && len(clauses) == 0 {
		return nil, fmt.Errorf("targets named no ids")
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Mode != out[j].Mode {
			return out[i].Mode < out[j].Mode
		}
		return out[i].Slot < out[j].Slot
	})
	// A single-slot pick_target prompt reads `target`, every other
	// shape reads `targets` (legal/choices.go).
	if t, ok := p["target"]; ok && t != nil {
		if len(out) != 1 {
			return nil, fmt.Errorf("this prompt takes exactly one target")
		}
		p["target"] = out[0]
		delete(p, "targets")
	} else {
		p["targets"] = out
	}
	return json.Marshal(p)
}

func inClause(c protocol.LegalTargetsView, id string) bool {
	for _, x := range c.Players {
		if x == id {
			return true
		}
	}
	for _, x := range c.Cards {
		if x == id {
			return true
		}
	}
	return false
}
