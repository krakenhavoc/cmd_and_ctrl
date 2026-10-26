package game

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// ordered_picks_keyed.go — ADR 0175's ordered collector with its state as
// DATA (#4290, ADR 0175 amendment of 2026-10-26), so a table waiting on
// any seat's pick of "each player sacrifices a creature of their choice"
// is a restore point.
//
// The closure form (ordered_permanent_choices.go) asks one seat at a time
// with its remaining legs, the picks so far, the objects' epochs and the
// card's continuation all held in the closure of the prompt's Then. Every
// one of those is data: the legs are seats with candidate refs (an ID and
// the ObjectEpoch it had when the collection began, CR 400.7) and bounds;
// the picks so far are seats with refs; the continuation is a registered
// run key (RegisterPromptRunThen) and the plain data the card put on it
// (RunData). They ride the prompt (PendingChoice.ordered), which is
// written to a restore point and cloned for an undo; the prompt is a keyed
// choose_cards pick (orderedPickLeg) whose registered continuation files
// the answer and asks the next seat.
//
// Everything else is the closure form's: seats choose one at a time, each
// seeing the earlier choices (CR 101.4b); nothing moves while choosing; a
// seat that has left, or whose candidates are all gone, is skipped; a leg
// whose only valid set is every candidate needs no prompt; picks are bound
// to the original objects, so a permanent that left and came back is not
// part of the final action. With Sacrifice the final action is one
// simultaneous sacrifice (SacrificeAllThenForEffect), and the
// continuation is handed the permanents that really left.

// orderedPicks is a keyed ordered collection in flight. A copy rides each
// prompt it asks, so the prompt is all a restore needs.
type orderedPicks struct {
	// Key is the registered run continuation (RegisterPromptRunThen), or ""
	// when nothing waits on the collection (a fire-and-forget sacrifice).
	Key    string    `json:"key,omitempty"`
	Source uuid.UUID `json:"source,omitempty"`
	// Question is each leg's header.
	Question string `json:"question,omitempty"`
	// Sacrifice ends the collection with one simultaneous sacrifice of
	// everything chosen, and hands the continuation what really left.
	Sacrifice bool `json:"sacrifice,omitempty"`
	// Destination is each leg's ChooseDestination.
	Destination ChooseDestination `json:"destination,omitempty"`
	// Carry, Nums and Asked are the run's plain data (RunData).
	Carry []uuid.UUID `json:"carry,omitempty"`
	Nums  []int       `json:"nums,omitempty"`
	Asked *AskedItem  `json:"asked,omitempty"`
	// Legs are every seat to ask, in order; Next is the one the prompt this
	// rides on is asking.
	Legs []orderedLeg `json:"legs"`
	Next int          `json:"next,omitempty"`
	// Picks are the seats already asked, with what each chose, bound to
	// the objects as they were. Earlier is what the chooser of this prompt
	// is shown of them: the picks still on the battlefield as the same
	// objects when the prompt went up (CR 101.4b).
	Picks   []orderedPick `json:"picks,omitempty"`
	Earlier []OfferPick   `json:"earlier,omitempty"`
}

// orderedLeg is one seat's turn: its candidates, each with the epoch it had
// when the collection began, and the bounds.
type orderedLeg struct {
	Seat  uuid.UUID   `json:"seat"`
	Cards []ObjectRef `json:"cards"`
	Min   int         `json:"min,omitempty"`
	Max   int         `json:"max,omitempty"`
	// Key is the leg's own registered continuation, which holds its set
	// rule (RegisterOrderedPickLegRule); empty for the engine's own, which
	// has none.
	Key string `json:"key,omitempty"`
}

// orderedPick is one seat's answer.
type orderedPick struct {
	Seat  uuid.UUID   `json:"seat"`
	Cards []ObjectRef `json:"cards,omitempty"`
}

func cloneOrderedPicks(o *orderedPicks) *orderedPicks {
	if o == nil {
		return nil
	}
	out := *o
	out.Carry = copyUUIDs(o.Carry)
	out.Nums = copyInts(o.Nums)
	out.Asked = copyAskedItem(o.Asked)
	if len(o.Legs) > 0 {
		out.Legs = make([]orderedLeg, len(o.Legs))
		for i, l := range o.Legs {
			out.Legs[i] = orderedLeg{Seat: l.Seat, Cards: cloneObjectRefs(l.Cards), Min: l.Min, Max: l.Max, Key: l.Key}
		}
	}
	if len(o.Picks) > 0 {
		out.Picks = make([]orderedPick, len(o.Picks))
		for i, p := range o.Picks {
			out.Picks[i] = orderedPick{Seat: p.Seat, Cards: cloneObjectRefs(p.Cards)}
		}
	}
	if len(o.Earlier) > 0 {
		out.Earlier = make([]OfferPick, len(o.Earlier))
		for i, p := range o.Earlier {
			out.Earlier[i] = OfferPick{Seat: p.Seat, Cards: copyUUIDs(p.Cards)}
		}
	}
	return &out
}

// orderedPickLegKey is the key of the engine's own continuation for a leg.
const orderedPickLegKey = "choose-cards/ordered-pick-leg"

// The continuation of every leg of a keyed ordered collection that has no
// set rule of its own. Registered in init so it is a boot-time key.
func init() { RegisterChooseCardsThen(orderedPickLegKey, orderedPickLegFn) }

// RegisterOrderedPickLegRule registers the continuation of a keyed ordered
// collection's leg that has a set rule over what its seat picks (Liliana,
// Dreadhorde General's "a permanent of each permanent type"), under its own
// key. The key is an on-disk identity, ledgered as "choose <key>". The rule
// is never called for an empty pick, is a function of the cards alone, and
// must capture nothing.
func RegisterOrderedPickLegRule(key string, rule SetRuleFunc) ChooseCardsThen {
	orderedLegKeys.Store(key, true)
	return RegisterChooseCardsThenRule(key, rule, orderedPickLegFn)
}

// orderedLegKeys are the keys RegisterOrderedPickLegRule made, so a
// collection can refuse a leg that names any other continuation.
var orderedLegKeys sync.Map

// orderedPickLegFn files the seat's answer and asks the next. A leg whose
// chooser left with the prompt open chose nothing (CR 800.4a).
func orderedPickLegFn(g *Game, r CardsChosen) error {
	st := cloneOrderedPicks(r.ordered)
	if st == nil || st.Next < 0 || st.Next >= len(st.Legs) {
		return fmt.Errorf("%w: an ordered pick with no collection", ErrInvalidParam)
	}
	leg := st.Legs[st.Next]
	pick := orderedPick{Seat: leg.Seat}
	if r.Answered {
		for _, id := range r.Picked {
			for _, ref := range leg.Cards {
				if ref.ID == id {
					pick.Cards = append(pick.Cards, ref)
					break
				}
			}
		}
	}
	st.Picks = append(st.Picks, pick)
	st.Next++
	return g.askOrderedPicksLocked(st)
}

// ChoosePermanentsInOrderSacrificeForEffect collects the legs one seat at a
// time and then sacrifices everything chosen as one simultaneous exit, with
// nothing waiting on it: ChoosePermanentsInOrderThenForEffect with a
// sacrifice for its end and the whole collection plain data (#4290, ADR
// 0175 amendment). The legs' Destination should be ChooseSacrifice.
//
// Caller must hold g.mu.
func (g *Game) ChoosePermanentsInOrderSacrificeForEffect(source uuid.UUID, legs []ChooseCardsPrompt) error {
	return g.startOrderedPicksLocked(legs, orderedPicks{Source: source, Sacrifice: true})
}

// ChoosePermanentsInOrderKeyedForEffect is ChoosePermanentsInOrderThenForEffect
// with the rest of the card as a registered run continuation
// (RegisterPromptRunThen) and `data` the plain data it is handed (#4290):
// the continuation reads RunDone.Picks(). The caller supplies the legs in
// APNAP order. Leg-level Then, Validate and Zone are rejected.
//
// Caller must hold g.mu.
func (g *Game) ChoosePermanentsInOrderKeyedForEffect(legs []ChooseCardsPrompt, then PromptRunThen, source uuid.UUID, data RunData) error {
	if then.key == "" {
		return fmt.Errorf("%w: keyed ordered pick with no continuation", ErrInvalidParam)
	}
	return g.startOrderedPicksLocked(legs, orderedPicks{Key: then.key, Source: source, Carry: data.Carry, Nums: data.Nums, Asked: data.Asked})
}

// startOrderedPicksLocked freezes the legs' candidates with their epochs
// and asks the first seat. Caller must hold g.mu.
func (g *Game) startOrderedPicksLocked(legs []ChooseCardsPrompt, st orderedPicks) error {
	st.Carry, st.Nums, st.Asked = copyUUIDs(st.Carry), copyInts(st.Nums), copyAskedItem(st.Asked)
	for i, leg := range legs {
		if leg.Then != nil || leg.Validate != nil || (leg.Zone != "" && leg.Zone != ZoneBattlefield) {
			return ErrInvalidParam
		}
		if leg.ThenKey.key != "" {
			if _, ok := orderedLegKeys.Load(leg.ThenKey.key); !ok {
				return ErrInvalidParam
			}
		}
		if i == 0 {
			st.Question, st.Destination = leg.Question, leg.Destination
		}
		ol := orderedLeg{Seat: leg.Chooser, Min: leg.Min, Max: leg.Max, Key: leg.ThenKey.key}
		seen := make(map[uuid.UUID]bool, len(leg.Cards))
		for _, id := range leg.Cards {
			if seen[id] {
				return ErrInvalidParam
			}
			seen[id] = true
			if c := findBattlefieldCard(g, id); c != nil {
				ol.Cards = append(ol.Cards, ObjectRef{ID: id, Epoch: c.ObjectEpoch})
			}
		}
		st.Legs = append(st.Legs, ol)
	}
	return g.askOrderedPicksLocked(&st)
}

// livePickRefs are the refs in `refs` that are still on the battlefield as
// the same objects.
func (g *Game) livePickRefs(refs []ObjectRef) []ObjectRef {
	var out []ObjectRef
	for _, ref := range refs {
		if c := findBattlefieldCard(g, ref.ID); c != nil && c.ObjectEpoch == ref.Epoch {
			out = append(out, ref)
		}
	}
	return out
}

// orderedPicksSoFarLocked is the picks as SeatCards, each card kept only
// while it is the same object it was chosen as.
func (g *Game) orderedPicksSoFarLocked(st *orderedPicks) PromptedPicks {
	out := make(PromptedPicks, 0, len(st.Picks))
	for _, p := range st.Picks {
		entry := SeatCards{Seat: p.Seat}
		for _, ref := range g.livePickRefs(p.Cards) {
			entry.Cards = append(entry.Cards, ref.ID)
		}
		out = append(out, entry)
	}
	return out
}

// askOrderedPicksLocked asks the first seat at or after st.Next that has
// something to choose, and finishes the collection when there is none.
// Caller must hold g.mu.
func (g *Game) askOrderedPicksLocked(st *orderedPicks) error {
	for ; st.Next < len(st.Legs); st.Next++ {
		leg := st.Legs[st.Next]
		if pl := g.playerByIDLocked(leg.Seat); pl == nil || pl.Eliminated {
			continue
		}
		live := g.livePickRefs(leg.Cards)
		if len(live) == 0 {
			continue
		}
		cards := make([]uuid.UUID, len(live))
		for i, ref := range live {
			cards[i] = ref.ID
		}
		hi := leg.Max
		if hi <= 0 || hi > len(cards) {
			hi = len(cards)
		}
		lo := max(0, min(leg.Min, hi))
		if lo == len(cards) && hi == len(cards) {
			// The only valid set is every candidate: no prompt, a click
			// the player had no choice about.
			st.Picks = append(st.Picks, orderedPick{Seat: leg.Seat, Cards: cloneObjectRefs(live)})
			continue
		}
		legKey := leg.Key
		if legKey == "" {
			legKey = orderedPickLegKey
		}
		earlier := g.orderedPicksSoFarLocked(st)
		leg1 := cloneOrderedPicks(st)
		leg1.Earlier = nil
		for _, e := range earlier {
			leg1.Earlier = append(leg1.Earlier, OfferPick{Seat: e.Seat, Cards: copyUUIDs(e.Cards)})
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:              PendingChoiceChooseCards,
			Chooser:           leg.Seat,
			FromPlayer:        leg.Seat,
			Count:             hi,
			Source:            st.Source,
			Reason:            st.Question,
			ChooseCards:       cards,
			ChooseMin:         lo,
			ChooseMax:         hi,
			ChooseDestination: st.Destination,
			ChooseLiveZone:    ZoneBattlefield,
			PickThen:          legKey,
			ordered:           leg1,
		})
		return nil
	}
	return g.finishOrderedPicksLocked(st)
}

// finishOrderedPicksLocked runs the end of the collection: the picks to
// the registered continuation, or, with Sacrifice, one simultaneous
// sacrifice of everything chosen and then the continuation with what
// really left. Caller must hold g.mu.
func (g *Game) finishOrderedPicksLocked(st *orderedPicks) error {
	picks := g.orderedPicksSoFarLocked(st)
	key, source := st.Key, st.Source
	data := RunData{Carry: copyUUIDs(st.Carry), Nums: copyInts(st.Nums), Asked: copyAskedItem(st.Asked)}
	run := func(g *Game, answer []SeatCards) error {
		if key == "" {
			return nil
		}
		fn, ok := lookupPromptRunThen(key)
		if !ok {
			return fmt.Errorf("%w: prompted-run continuation %q", ErrUnknownEffectKey, key)
		}
		return fn(g, RunDone{Source: source, Landed: answer, Carry: data.Carry, Nums: data.Nums, Asked: data.Asked})
	}
	if !st.Sacrifice {
		return run(g, picks)
	}
	finish := func(g *Game, sacrificed []uuid.UUID) error {
		landed := make(map[uuid.UUID]bool, len(sacrificed))
		for _, id := range sacrificed {
			landed[id] = true
		}
		out := make([]SeatCards, 0, len(picks))
		for _, pick := range picks {
			entry := SeatCards{Seat: pick.Seat}
			for _, id := range pick.Cards {
				if landed[id] {
					entry.Cards = append(entry.Cards, id)
				}
			}
			out = append(out, entry)
		}
		return run(g, out)
	}
	chosen := picks.Cards()
	if len(chosen) == 0 {
		return finish(g, nil)
	}
	return g.SacrificeAllThenForEffect(source, chosen, finish)
}

// unknownOrderedPicksKeys lists the continuation an ordered collection on
// disk names, when this binary does not have it.
func unknownOrderedPicksKeys(o *orderedPicks) []string {
	if o == nil || o.Key == "" || KnownPromptRunThen(o.Key) {
		return nil
	}
	return []string{"prompted-run continuation " + o.Key}
}
