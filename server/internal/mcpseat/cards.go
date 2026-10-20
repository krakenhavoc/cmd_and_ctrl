package mcpseat

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// cards.go — #4186: name exactly which cards a choose-cards-from-hand
// prompt takes, instead of finding the set among the enumerator's capped
// combinations. With 13 cards in hand and a hand size of 7, a cleanup
// discard has C(13,6) = 1,716 answers and the list stops at 512, so most
// sets could not be chosen at all.
//
// Nothing here needs a new frame either. Every such move already carries
// its pick as `card_ids` in its params (discard_selection, and
// resolve_choice for a discard_from_hand or revealed_hand_pick prompt),
// and the server checks whatever set it is sent: the count and that every
// card is in the hand (Game.DiscardSelection), or among the prompt's
// options (the choice resolver). The seat reads the pool and the count
// from the view, checks the agent's set against them, and swaps it into
// the chosen move's params.

// cardPick is what a choose-cards move picks from: the pool of ids it may
// name, the sizes a set may have, and the prompt's key, so two moves of
// one prompt share a footer line.
type cardPick struct {
	key   string
	what  string // "discard to hand size", or the prompt's reason
	pool  []string
	sizes []int
}

// moveCardPick reports the pool and sizes of a choose-cards move, or false
// for a move whose cards cannot be named.
func moveCardPick(v *protocol.GameView, me string, m legal.Move) (cardPick, bool) {
	if v == nil {
		return cardPick{}, false
	}
	switch m.Type {
	case legal.TypeDiscardSelection:
		n, ok := v.DiscardPending[me]
		self := mySeat(v, me)
		if !ok || n <= 0 || self == nil {
			return cardPick{}, false
		}
		var pool []string
		for i := range self.Hand.Cards {
			pool = append(pool, self.Hand.Cards[i].InstanceID)
		}
		return cardPick{key: legal.CleanupDiscardChoice, what: "discard to hand size", pool: pool, sizes: []int{n}}, true
	case legal.TypeResolveChoice:
		var p struct {
			ChoiceID string   `json:"choice_id"`
			CardIDs  []string `json:"card_ids"`
		}
		if len(m.Params) == 0 || json.Unmarshal(m.Params, &p) != nil || p.ChoiceID == "" || p.CardIDs == nil {
			return cardPick{}, false
		}
		for i := range v.PendingChoices {
			c := &v.PendingChoices[i]
			if c.ID != p.ChoiceID || c.Chooser != me {
				continue
			}
			what := c.Reason
			if what == "" {
				what = c.Kind
			}
			switch c.Kind {
			case "discard_from_hand":
				// ADR 0116 §7: Eligible is what the card lets you
				// choose; absent, every option.
				pool := c.Eligible
				if pool == nil {
					for j := range c.Options {
						pool = append(pool, c.Options[j].InstanceID)
					}
				}
				if c.Count <= 0 {
					return cardPick{}, false
				}
				return cardPick{key: c.ID, what: what, pool: pool, sizes: []int{c.Count}}, true
			case "revealed_hand_pick":
				if c.ChooseMax <= 0 {
					return cardPick{}, false
				}
				sizes := []int{c.ChooseMax}
				if c.ChooseMin == 0 {
					sizes = []int{0, c.ChooseMax}
				}
				return cardPick{key: c.ID, what: what, pool: c.Eligible, sizes: sizes}, true
			}
		}
	}
	return cardPick{}, false
}

// sizesText says how many cards a pick takes: "6", or "0 or 1".
func (p cardPick) sizesText() string {
	s := make([]string, len(p.sizes))
	for i, n := range p.sizes {
		s[i] = strconv.Itoa(n)
	}
	return strings.Join(s, " or ")
}

// applyCards swaps the agent's set into a choose-cards move's card_ids,
// after checking it against the pool and the count.
func applyCards(v *protocol.GameView, me string, m legal.Move, params json.RawMessage, ids []string) (json.RawMessage, error) {
	pick, ok := moveCardPick(v, me, m)
	if !ok {
		return nil, fmt.Errorf("this move does not choose cards from a hand: send it without cards")
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, fmt.Errorf("an empty card id")
		}
		if !slices.Contains(pick.pool, id) {
			return nil, fmt.Errorf("%s is not one of the cards this prompt lets you choose (the move list's CHOOSE line names them)", id)
		}
		if seen[id] {
			return nil, fmt.Errorf("%s is named twice", id)
		}
		seen[id] = true
		out = append(out, id)
	}
	if !slices.Contains(pick.sizes, len(out)) {
		return nil, fmt.Errorf("this prompt takes %s cards, not %d", pick.sizesText(), len(out))
	}
	p := map[string]any{}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
	}
	p["card_ids"] = out
	return json.Marshal(p)
}

// cardPickLines is the move list's footer for each choose-cards prompt in
// it: how many cards, which ones, and how to name them with act.
func cardPickLines(v *protocol.GameView, w *window, shown map[int]bool) string {
	type entry struct {
		pick cardPick
		move int
	}
	var order []string
	byKey := map[string]entry{}
	for i, m := range w.moves {
		if !shown[i] {
			continue
		}
		p, ok := moveCardPick(v, w.me, m)
		if !ok {
			continue
		}
		if _, dup := byKey[p.key]; dup {
			continue
		}
		byKey[p.key] = entry{p, i}
		order = append(order, p.key)
	}
	sort.Strings(order)
	var b strings.Builder
	for _, k := range order {
		e := byKey[k]
		names := make([]string, 0, len(e.pick.pool))
		for _, id := range e.pick.pool {
			name := id
			if c, _ := findCard(v, id); c != nil {
				name = boardtext.CardName(c) + " [" + id + "]"
			}
			names = append(names, name)
		}
		fmt.Fprintf(&b, "CHOOSE %s cards (%s) from: %s\n", e.pick.sizesText(), cut(e.pick.what, 80), strings.Join(names, ", "))
		fmt.Fprintf(&b, "  Any set is legal, listed or not: act(window: %q, move: %d, cards: [\"<id>\", …]) sends exactly those cards.\n", w.token, e.move)
	}
	return b.String()
}
