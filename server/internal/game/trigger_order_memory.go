package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// trigger_order_memory.go — "use this order again" for a repeated batch
// (ADR 0018's eighth amendment, #3016).
//
// A seat that orders the same triggers again and again may keep the
// order: the trigger-order prompt's "Use this order again" toggle sends
// `remember: true`, and the answer's ranking of the batch's rows is
// filed under the prompt's key. The next prompt with the same key is an
// ADR 0127 automatic answer (auto_answer.go): the room answers it as a
// commit of its own, with an "(automatic)" log line, a free undo and an
// "Ask me next time" notice, and it does not restart the CR 732 loop
// run.
//
//   - A ROW is one triggered ability as the catalog knows it: its
//     AbilityRef, or for an engine keyword trigger its body plus its
//     source's catalog key (A1).
//   - The batch KEY is the sorted set of the batch's distinct rows:
//     counts and source instances are ignored (A1). An item with no row,
//     or whose source is face down or in a hidden zone (G), leaves the
//     whole batch unkeyed, and it is asked as before.
//   - What is remembered is a RANKING of the rows, applied by a stable
//     sort, so copies of one row stay together in the order they
//     triggered (B). An answer that interleaves copies of one row is not
//     a ranking and cannot be remembered.
//   - The memory is game state (C1, H): per seat, per game, at most
//     MaxTriggerOrderMemory entries, cloned and snapshotted, and undone
//     with the commit that taught it.

// MaxTriggerOrderMemory bounds a seat's remembered orders (owner
// decision 7). A new one past it replaces the one applied least
// recently.
const MaxTriggerOrderMemory = 32

// TriggerOrderAutoAnswerPrefix starts every trigger-order prompt's
// AutoAnswerKey; the rest is the batch key.
const TriggerOrderAutoAnswerPrefix = "trigger_order|"

// MaxForgetTriggerOrderKeyLen bounds the key forget_trigger_order takes,
// in bytes. A real key is a few rows of about 80 bytes each.
const MaxForgetTriggerOrderKeyLen = 8192

// triggerOrderKeySep joins the rows of a batch key.
const triggerOrderKeySep = "+"

// AutoAnswerCallOrder is the Event.Call of an EventAutoAnswer that
// applied a remembered trigger order. Event.Label is then the rows'
// stack labels in resolution order, joined by TriggerOrderLabelJoin.
const AutoAnswerCallOrder = "order"

// TriggerOrderLabelJoin joins the labels of an applied order.
const TriggerOrderLabelJoin = ", then "

// AskedByHandChanged: the batch no longer matches its remembered order
// (an item changed rows since the prompt was queued).
const AskedByHandChanged AskedByHand = "changed"

// AskedByHandHidden: a source in the batch is face down or in a hidden
// zone, so applying the order would name it (ADR 0018 eighth
// amendment G, CR 708.2).
const AskedByHandHidden AskedByHand = "hidden"

// RememberedTriggerOrder is one remembered ranking.
type RememberedTriggerOrder struct {
	// Rows are the batch's row identities in resolution order: the
	// first resolves first.
	Rows []string `json:"rows"`
	// Used is a recency stamp, higher for an order taught or applied
	// more recently. The order applied least recently is the one a new
	// order past the cap replaces.
	Used int `json:"used"`
}

// copyTriggerOrderMemory deep-copies a seat's memory; nil for none.
func copyTriggerOrderMemory(in map[string]RememberedTriggerOrder) map[string]RememberedTriggerOrder {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]RememberedTriggerOrder, len(in))
	for k, v := range in {
		out[k] = RememberedTriggerOrder{Rows: append([]string(nil), v.Rows...), Used: v.Used}
	}
	return out
}

// keywordTriggerRowBodies are the bodies of the engine keyword triggers
// (prowess, exalted and the rest of NewKeyedTriggeredItem's keyword
// callers). Such an item has no AbilityRef, and its row is the body plus
// its source's catalog key. A delayed or reflexive trigger also carries
// a body, but it is not a row of anything the player can see, so it is
// not listed and leaves its batch unkeyed.
//
// Filled in init: the bodies' own initialisers reach the trigger drain,
// which reads this set, so a package-level initialiser would be a cycle.
var keywordTriggerRowBodies map[string]bool

func init() {
	keywordTriggerRowBodies = map[string]bool{
		prowessPumpBody.key:          true,
		exaltedPumpBody.key:          true,
		annihilatorSacrificeBody.key: true,
		undyingReturnBody.key:        true,
		persistReturnBody.key:        true,
		decayedScheduleBody.key:      true,
		modularMoveCountersBody.key:  true,
		evolveGrowBody.key:           true,
		renownGrowBody.key:           true,
		casualtyCopyBody.key:         true,
		madnessOfferBody.key:         true,
		miracleOfferBody.key:         true,
		suspendTickBody.key:          true,
		suspendFreeCastBody.key:      true,
	}
}

// triggerRowLocked is t's row identity, and whether it has one: its
// AbilityRef as `key|slot|ref|name`, or for an engine keyword trigger
// `<body>@<source catalog key>`. hidden is true when the item's source
// is face down or in a library or a hand, which leaves it without one
// (G). Caller must hold g.mu.
func (g *Game) triggerRowLocked(t *StackItem) (row string, hidden, ok bool) {
	if t == nil || t.SourceCardID == uuid.Nil {
		return "", false, false
	}
	var source *Card
	if z, i := g.locateCardLocked(t.SourceCardID); z != nil && i >= 0 && i < len(z.Cards) {
		if z.Kind == ZoneLibrary || z.Kind == ZoneHand {
			return "", true, false
		}
		source = &z.Cards[i]
		if source.FaceDown {
			return "", true, false
		}
	}
	if ab := t.Params.Ability; ab != nil && ab.Key != "" {
		if ab.Slot != AbilitySlotTriggered {
			return "", false, false
		}
		return ab.Key + "|" + ab.Slot + "|" + ab.Ref + "|" + ab.Name, false, true
	}
	if t.Body == "" || !keywordTriggerRowBodies[t.Body] || source == nil {
		return "", false, false
	}
	key := catalogKeyOf(source)
	if key == "" {
		return "", false, false
	}
	return t.Body + "@" + key, false, true
}

// triggerOrderBatchKeyLocked is the AutoAnswerKey of a trigger-order
// prompt over items: TriggerOrderAutoAnswerPrefix and the sorted set of
// their distinct rows. Empty when any item has no row; hidden says one
// of them was left out for a hidden source. Caller must hold g.mu.
func (g *Game) triggerOrderBatchKeyLocked(items []*StackItem) (key string, hidden bool) {
	if len(items) == 0 {
		return "", false
	}
	seen := make(map[string]bool, len(items))
	rows := make([]string, 0, len(items))
	for _, t := range items {
		row, h, ok := g.triggerRowLocked(t)
		if !ok {
			return "", h
		}
		if !seen[row] {
			seen[row] = true
			rows = append(rows, row)
		}
	}
	sort.Strings(rows)
	return TriggerOrderAutoAnswerPrefix + strings.Join(rows, triggerOrderKeySep), false
}

// pendingTriggersByIDLocked returns the pending triggers named by ids,
// in ids' order, and false if any of them is no longer pending. Caller
// must hold g.mu.
func (g *Game) pendingTriggersByIDLocked(ids []uuid.UUID) ([]*StackItem, bool) {
	byID := make(map[uuid.UUID]*StackItem, len(g.PendingTriggers))
	for _, t := range g.PendingTriggers {
		if t != nil {
			byID[t.ID] = t
		}
	}
	out := make([]*StackItem, 0, len(ids))
	for _, id := range ids {
		t, ok := byID[id]
		if !ok {
			return nil, false
		}
		out = append(out, t)
	}
	return out, true
}

// triggerOrderRankingLocked is the ranking an answer gives: the rows of
// order (resolution order) with each row's copies together, and false
// when the answer interleaves copies of one row (B).
func (g *Game) triggerOrderRankingLocked(order []*StackItem) ([]string, bool) {
	rows := make([]string, 0, len(order))
	done := map[string]bool{}
	for _, t := range order {
		row, _, ok := g.triggerRowLocked(t)
		if !ok {
			return nil, false
		}
		if len(rows) > 0 && rows[len(rows)-1] == row {
			continue
		}
		if done[row] {
			return nil, false
		}
		done[row] = true
		rows = append(rows, row)
	}
	return rows, true
}

// TriggerOrderRowsLocked is the row index (1-based, in first-appearance
// order) of each item of a trigger-order prompt, for the client to tell
// which orders keep each row's copies together; nil when the prompt has
// no key. Caller must hold g.mu (the protocol projection runs under it).
func (g *Game) TriggerOrderRowsLocked(c *PendingChoice) []int {
	if c == nil || c.Kind != PendingChoiceTriggerOrder || c.AutoAnswerKey == "" {
		return nil
	}
	items, ok := g.pendingTriggersByIDLocked(c.TriggerOrderIDs)
	if !ok {
		return nil
	}
	index := map[string]int{}
	out := make([]int, len(items))
	for i, t := range items {
		row, _, ok := g.triggerRowLocked(t)
		if !ok {
			return nil
		}
		if index[row] == 0 {
			index[row] = len(index) + 1
		}
		out[i] = index[row]
	}
	return out
}

// TriggerOrderRememberedLocked reports whether c's chooser has an order
// remembered for c's batch. Caller must hold g.mu (the protocol
// projection runs under it).
func (g *Game) TriggerOrderRememberedLocked(c *PendingChoice) bool {
	if c == nil || c.Kind != PendingChoiceTriggerOrder || c.AutoAnswerKey == "" {
		return false
	}
	p := g.playerByIDLocked(c.Chooser)
	if p == nil {
		return false
	}
	_, ok := p.TriggerOrderMemory[c.AutoAnswerKey]
	return ok
}

// nextTriggerOrderStamp is the recency stamp a write gives an
// entry: one past the seat's newest.
func nextTriggerOrderStamp(mem map[string]RememberedTriggerOrder) int {
	n := 0
	for _, v := range mem {
		if v.Used > n {
			n = v.Used
		}
	}
	return n + 1
}

// rememberTriggerOrderLocked applies a trigger-order answer's
// `remember` to the chooser's memory (D1): true files the answer's
// ranking under the prompt's key, false forgets any order filed there.
// It refuses, writing nothing, a `remember: true` on a prompt with no
// key or an answer that is not a ranking. A bot seat, and a chooser
// another player controls (CR 723.5, the order is not theirs to teach),
// are left as they are. Caller must hold g.mu.
func (g *Game) rememberTriggerOrderLocked(c *PendingChoice, order []uuid.UUID, remember bool) error {
	p := g.playerByIDLocked(c.Chooser)
	if p == nil || p.IsBot || g.ControllerOfPlayerLocked(c.Chooser) != c.Chooser {
		return nil
	}
	if !remember {
		if c.AutoAnswerKey != "" {
			delete(p.TriggerOrderMemory, c.AutoAnswerKey)
			if len(p.TriggerOrderMemory) == 0 {
				p.TriggerOrderMemory = nil
			}
		}
		return nil
	}
	if c.AutoAnswerKey == "" {
		return fmt.Errorf("%w: this trigger order can't be reused", ErrInvalidParam)
	}
	items, ok := g.pendingTriggersByIDLocked(order)
	if !ok {
		return fmt.Errorf("%w: this trigger order can't be reused", ErrInvalidParam)
	}
	rows, ok := g.triggerOrderRankingLocked(items)
	if !ok {
		return fmt.Errorf("%w: only an order that keeps each card's triggers together can be reused", ErrInvalidParam)
	}
	if key, _ := g.triggerOrderBatchKeyLocked(items); key != c.AutoAnswerKey {
		return fmt.Errorf("%w: this trigger order can't be reused", ErrInvalidParam)
	}
	g.storeTriggerOrderLocked(p, c.AutoAnswerKey, rows)
	return nil
}

// storeTriggerOrderLocked files rows under key, stamped as the newest,
// and past the cap drops the entry applied least recently (ties by
// key, so every replica drops the same one). Caller must hold g.mu.
func (g *Game) storeTriggerOrderLocked(p *Player, key string, rows []string) {
	if p.TriggerOrderMemory == nil {
		p.TriggerOrderMemory = map[string]RememberedTriggerOrder{}
	}
	stamp := nextTriggerOrderStamp(p.TriggerOrderMemory)
	if _, exists := p.TriggerOrderMemory[key]; !exists {
		for len(p.TriggerOrderMemory) >= MaxTriggerOrderMemory {
			oldest, oldestUsed := "", 0
			for k, v := range p.TriggerOrderMemory {
				if oldest == "" || v.Used < oldestUsed || (v.Used == oldestUsed && k < oldest) {
					oldest, oldestUsed = k, v.Used
				}
			}
			delete(p.TriggerOrderMemory, oldest)
		}
	}
	p.TriggerOrderMemory[key] = RememberedTriggerOrder{Rows: append([]string(nil), rows...), Used: stamp}
}

// triggerOrderHandReasonLocked is why c, a trigger-order prompt whose
// chooser has an order remembered for it, must be asked by hand (F
// checks 1 and 2), or "" when the order may answer it. Caller must hold
// g.mu.
func (g *Game) triggerOrderHandReasonLocked(c *PendingChoice) AskedByHand {
	items, ok := g.pendingTriggersByIDLocked(c.TriggerOrderIDs)
	if !ok {
		return AskedByHandChanged
	}
	key, hidden := g.triggerOrderBatchKeyLocked(items)
	if hidden {
		return AskedByHandHidden
	}
	if key != c.AutoAnswerKey {
		return AskedByHandChanged
	}
	return ""
}

// rememberedTriggerOrderLocked is the answer the chooser's remembered
// order gives c: c's items stably sorted by their row's rank, and their
// stack labels in that order, one per row. ok is false when the memory
// no longer covers every row. Caller must hold g.mu.
func (g *Game) rememberedTriggerOrderLocked(c *PendingChoice, mem RememberedTriggerOrder) (order []uuid.UUID, labels []string, ok bool) {
	items, ok := g.pendingTriggersByIDLocked(c.TriggerOrderIDs)
	if !ok {
		return nil, nil, false
	}
	rank := make(map[string]int, len(mem.Rows))
	for i, r := range mem.Rows {
		rank[r] = i
	}
	type ranked struct {
		item *StackItem
		row  string
		rank int
	}
	list := make([]ranked, len(items))
	for i, t := range items {
		row, _, ok := g.triggerRowLocked(t)
		if !ok {
			return nil, nil, false
		}
		r, known := rank[row]
		if !known {
			return nil, nil, false
		}
		list[i] = ranked{item: t, row: row, rank: r}
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].rank < list[j].rank })
	order = make([]uuid.UUID, len(list))
	for i, e := range list {
		order[i] = e.item.ID
		if i == 0 || list[i-1].row != e.row {
			labels = append(labels, e.item.Label)
		}
	}
	return order, labels, true
}

// ForgetTriggerOrder forgets playerID's remembered trigger order filed
// under key, or every one of them when key is empty ("Ask me next time"
// and the game menu's "Forget remembered trigger orders"). A seat's own
// choice about its own memory: it emits no event and mints no undo
// entry (actions.MintsNoUndo). A key with nothing filed under it is not
// an error. Takes the write lock.
func (g *Game) ForgetTriggerOrder(playerID uuid.UUID, key string) error {
	if len(key) > MaxForgetTriggerOrderKeyLen {
		return ErrInvalidParam
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if key == "" {
		p.TriggerOrderMemory = nil
		return nil
	}
	delete(p.TriggerOrderMemory, key)
	if len(p.TriggerOrderMemory) == 0 {
		p.TriggerOrderMemory = nil
	}
	return nil
}

// TriggerOrderMemoryCountLocked is how many trigger orders playerID has
// remembered. Caller must hold g.mu (the protocol projection runs under
// it).
func (g *Game) TriggerOrderMemoryCountLocked(playerID uuid.UUID) int {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return 0
	}
	return len(p.TriggerOrderMemory)
}
