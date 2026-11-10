package game

import (
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// explore.go — the CR 701.44 keyword action "<permanent> explores"
// (#2720).
//
//	701.44a … that permanent's controller reveals the top card of their
//	library. If a land card is revealed this way, that player puts that
//	card into their hand. Otherwise, that player puts a +1/+1 counter on
//	the exploring permanent and may put the revealed card into their
//	graveyard.
//
// One verb, built from parts that already exist: the reveal is
// RevealTopOfLibraryForEffect, the land's trip to hand and the nonland's
// trip to the graveyard are routeCardToZoneLocked (so a commander in a
// library still gets its CR 903.9 question), the counter is
// AddCounterByKeyedForEffect (so Hardened Scales and friends settle
// before the question), and "may put it into your graveyard" is the
// general confirm prompt. It is not surveil: a "whenever you surveil"
// payoff must not fire on an explore, so the question is not a surveil
// prompt.
//
// CR 701.44b: the permanent has explored once the process is complete,
// even if some or all of it was impossible, so EventExplored is emitted
// at the end of every branch, an empty library and a departed explorer
// included. It is its own kind, as EventDiscover is, so an explore
// payoff ("whenever a creature you control explores") fires on nothing
// else.
//
// CR 701.44c: a permanent that has left the battlefield still explores
// as it last existed. The caller names the controller (the item that
// told it to explore knows it), the reveal and the land still happen,
// and only the counter has nowhere to go: the ObjectRef's epoch is
// checked, so a creature that was flickered in response is a new object
// and gets no counter (CR 400.7).
//
// CR 701.44d, several permanents told to explore at once, is
// ExploreEachForEffect (explore_each.go). The CR 614 replacements of the
// action are the two registered keys below: "instead you scry 1, then that
// creature explores" (RegisterExploreScryKey) and "instead it explores,
// then it explores again" (RegisterExploreAgainKey, #5425).

// EventExplored records that a permanent explored (CR 701.44b). CardID
// is the permanent that explored, Actor its controller, Source the card
// whose effect told it to explore.
const EventExplored EventKind = "explored"

// ExploreForEffect has the permanent `explorer` explore (CR 701.44a),
// controlled by `controller`, on behalf of `source`. There is no rest-of-
// the-effect to hand it (#4298): the graveyard question is a keyed confirm,
// and a closure on it would be a table that cannot be saved. A card that
// wants something after the explore watches EventExplored (CR 701.44b).
//
// Caller must hold g.mu.
func (g *Game) ExploreForEffect(source uuid.UUID, explorer ObjectRef, controller uuid.UUID) error {
	return g.exploreLocked(source, explorer, controller, exploreRest{})
}

// ExploreTimesForEffect has the permanent explore n times in a row (CR 701.44a,
// "explores X times"): each explore finishes, its question answered, before
// the next reveals, so a nonland card left on top is revealed again. n of 0 or
// less does nothing. The count rides the keyed tails and the prompt as data.
//
// Caller must hold g.mu.
func (g *Game) ExploreTimesForEffect(source uuid.UUID, explorer ObjectRef, controller uuid.UUID, n int) error {
	return g.ExploreTimesThenForEffect(source, explorer, controller, n, TailRef{})
}

// ExploreTimesThenForEffect is ExploreTimesForEffect and then the registered
// tail `then`, run once the last explore has finished (CR 701.44b) with the
// explorer as its Carry[0], the explore's controller as its Player and the
// source as its Source: "explores. Then you may ..." (Deepfathom Echo).
// The tail rides the explore's own prompts as data, so a table waiting on the
// graveyard question is still a restore point.
//
// Caller must hold g.mu.
func (g *Game) ExploreTimesThenForEffect(source uuid.UUID, explorer ObjectRef, controller uuid.UUID, n int, then TailRef) error {
	if n <= 0 {
		return nil
	}
	return g.exploreLocked(source, explorer, controller, exploreRest{More: n - 1, Then: then.Key()})
}

// exploreRest is what follows one explore: the explores a replacement
// already made of this one (Again, which no replacement sees a second time,
// CR 614.5), more explores of the same permanent (More, each a new
// instruction), the other permanents the same instruction told to explore
// (Queue, CR 701.44d), then an optional registered tail. Plain data, carried
// through the explore's tails and prompts as numbers and words.
type exploreRest struct {
	More  int
	Then  string
	Again int
	Queue []ObjectRef
}

// word is the rest as a prompt's one word: "" when nothing follows.
func (r exploreRest) word() string {
	if r.More <= 0 && r.Then == "" && r.Again <= 0 && len(r.Queue) == 0 {
		return ""
	}
	return strconv.Itoa(r.More) + "|" + r.tailWord()
}

// tailWord is everything but More, as the word of a tail whose N is More.
// A rest with no Again and no Queue is its tail's key alone, which is the
// word every snapshot written before #5425 holds.
func (r exploreRest) tailWord() string {
	if r.Again <= 0 && len(r.Queue) == 0 {
		return r.Then
	}
	refs := make([]string, len(r.Queue))
	for i, q := range r.Queue {
		refs[i] = q.ID.String() + ":" + strconv.Itoa(q.Epoch)
	}
	return r.Then + "|" + strconv.Itoa(r.Again) + "|" + strings.Join(refs, ",")
}

func exploreRestFromWord(w string) exploreRest {
	n, tail, _ := strings.Cut(w, "|")
	more, _ := strconv.Atoi(n)
	return exploreRestFromTail(more, tail)
}

// exploreRestFromTail reads a tail's N and Word back (tailWord).
func exploreRestFromTail(more int, w string) exploreRest {
	parts := strings.SplitN(w, "|", 3)
	r := exploreRest{More: more, Then: parts[0]}
	if len(parts) < 3 {
		return r
	}
	r.Again, _ = strconv.Atoi(parts[1])
	for _, ref := range strings.Split(parts[2], ",") {
		id, epoch, _ := strings.Cut(ref, ":")
		parsed, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		e, _ := strconv.Atoi(epoch)
		r.Queue = append(r.Queue, ObjectRef{ID: parsed, Epoch: e})
	}
	return r
}

// exploreLocked is one explore with `rest` still to follow it, through the
// explore replacements first.
//
// Caller must hold g.mu.
func (g *Game) exploreLocked(source uuid.UUID, explorer ObjectRef, controller uuid.UUID, rest exploreRest) error {
	// CR 614: "If a creature you control would explore, instead you scry 1,
	// then that creature explores" (Twists and Turns). One replacement per
	// copy, each applying once to what the one before it made (CR 614.5), so
	// N copies are N scries and then the explore.
	//
	// "If a creature you control would explore, instead it explores, then it
	// explores again" (Topography Tracker, #5425) is read here too, before
	// any scry: each copy doubles the explores this one instruction makes,
	// and the explores it adds are not replaced a second time (CR 614.5),
	// so they ride the rest as Again. The two replacements commute: either
	// order of applying them (CR 616.1) is the scries, then the explores.
	if !g.explorerIsACreatureLocked(explorer) {
		return g.exploreNowLocked(source, explorer, controller, rest)
	}
	rest.Again = g.exploreAgainCountLocked(controller)
	if n := g.exploreScryCountLocked(controller); n > 0 {
		tail := KeyedTail{
			Ref: exploreAfterScryTail, Player: controller, Source: source,
			Carry: []uuid.UUID{explorer.ID}, Nums: []int{explorer.Epoch}, N: n - 1, Word: rest.word(),
		}
		g.ScryKeyedForEffect(controller, source, 1, tail)
		return nil
	}
	return g.exploreNowLocked(source, explorer, controller, rest)
}

// exploreScryKeys are the catalog keys of the permanents that print "If a
// creature you control would explore, instead you scry 1, then that
// creature explores" (RegisterExploreScryKey). Written at init only.
var exploreScryKeys = struct {
	sync.RWMutex
	keys map[string]bool
}{keys: map[string]bool{}}

// RegisterExploreScryKey registers a card's catalog key as the explore
// replacement above. Call it once from the card's init.
func RegisterExploreScryKey(key string) {
	exploreScryKeys.Lock()
	defer exploreScryKeys.Unlock()
	exploreScryKeys.keys[key] = true
}

// exploreScryCountLocked is how many of those replacements `controller`'s
// permanents carry: through the catalog key that says the ability is still
// there, so a Twists and Turns that lost its abilities replaces nothing.
// Caller must hold g.mu.
func (g *Game) exploreScryCountLocked(controller uuid.UUID) int {
	exploreScryKeys.RLock()
	defer exploreScryKeys.RUnlock()
	return g.exploreKeyCountLocked(controller, exploreScryKeys.keys)
}

// exploreKeyCountLocked counts `controller`'s permanents whose catalog key
// is in keys. Caller must hold g.mu and the lock that guards keys.
func (g *Game) exploreKeyCountLocked(controller uuid.UUID, keys map[string]bool) int {
	if g.Battlefield == nil || len(keys) == 0 {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != controller {
			continue
		}
		if key := catalogAbilityKeyOf(c); key != "" && keys[key] {
			n++
		}
	}
	return n
}

// exploreAgainKeys are the catalog keys of the permanents that print "If a
// creature you control would explore, instead it explores, then it explores
// again" (RegisterExploreAgainKey). Written at init only.
var exploreAgainKeys = struct {
	sync.RWMutex
	keys map[string]bool
}{keys: map[string]bool{}}

// RegisterExploreAgainKey registers a card's catalog key as that
// replacement (#5425). Call it once from the card's init.
func RegisterExploreAgainKey(key string) {
	exploreAgainKeys.Lock()
	defer exploreAgainKeys.Unlock()
	exploreAgainKeys.keys[key] = true
}

// exploreAgainCountLocked is how many explores those replacements add to
// one instruction to explore: one copy makes it two explores, two copies
// four, each applying once to what the one before it made (CR 614.5). The
// answer is the explores beyond the first, capped as a doubled keyword
// action is (maxKeywordActionRepeats).
// Caller must hold g.mu.
func (g *Game) exploreAgainCountLocked(controller uuid.UUID) int {
	exploreAgainKeys.RLock()
	defer exploreAgainKeys.RUnlock()
	total := 1
	for n := g.exploreKeyCountLocked(controller, exploreAgainKeys.keys); n > 0 && total < maxKeywordActionRepeats; n-- {
		total *= 2
	}
	return total - 1
}

// explorerIsACreatureLocked is the "a creature you control" of both
// replacements: a permanent that is on the battlefield as the object the
// instruction named and is no longer a creature explores unreplaced. One
// that has left is read as it last existed (CR 701.44c), and every printed
// instruction to explore names a creature.
// Caller must hold g.mu.
func (g *Game) explorerIsACreatureLocked(explorer ObjectRef) bool {
	c := findBattlefieldCard(g, explorer.ID)
	return c == nil || c.ObjectEpoch != explorer.Epoch || c.IsCreature()
}

// exploreAfterScryTail is the explore once the replacement's scries are
// done: the next scry while any are owed (the tail's N), then the explore
// itself, of the object the ability named (Carry and Nums). An on-disk key:
// never renamed, never reused.
var exploreAfterScryTail = RegisterTail("engine/explore-after-scry", func(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) == 0 || len(o.Tail.Nums) == 0 {
		return nil
	}
	if o.Tail.N > 0 {
		next := o.Tail
		next.N--
		g.ScryKeyedForEffect(next.Player, next.Source, 1, next)
		return nil
	}
	return g.exploreNowLocked(o.Tail.Source, ObjectRef{ID: o.Tail.Carry[0], Epoch: o.Tail.Nums[0]}, o.Tail.Player, exploreRestFromWord(o.Tail.Word))
})

// exploreNowLocked is the explore itself (CR 701.44a), with no replacement
// left to apply, and `rest` to follow it.
//
// Caller must hold g.mu.
func (g *Game) exploreNowLocked(source uuid.UUID, explorer ObjectRef, controller uuid.UUID, rest exploreRest) error {
	done := func(g *Game) error { return g.finishExploreLocked(source, explorer.ID, controller, uuid.Nil, rest) }
	p := g.playerByIDLocked(controller)
	if p == nil || p.Eliminated {
		return done(g)
	}
	revealed := g.RevealTopOfLibraryForEffect(controller, source, 1, exploreReason(g, explorer))
	if len(revealed) == 0 {
		return done(g)
	}
	top, ok := g.cardInZoneLocked(p.Library, revealed[0])
	if !ok {
		return done(g)
	}
	if top.IsLand() {
		// #4361 S11b: the rest is a registered tail, so a CR 616 prompt
		// on the land's trip to hand is a restore point.
		return g.routeCardKeyedLocked(zoneRoute{
			CardID:   top.InstanceID,
			Dst:      ZoneHand,
			DstOwner: controller,
			Actor:    controller,
			Source:   source,
		}, exploreDone(source, explorer.ID, controller, top.InstanceID, rest))
	}
	// #4253 slice 7: the question is the counter's keyed tail (the explorer
	// and the revealed card ride it as data), so a CR 616 ordering prompt over
	// the +1/+1 counter is a restore point.
	ask := KeyedTail{Ref: exploreAskTail, Source: source, Player: controller, Carry: []uuid.UUID{explorer.ID, top.InstanceID}, N: rest.More, Word: rest.tailWord()}
	if perm := findBattlefieldCard(g, explorer.ID); perm != nil && perm.ObjectEpoch == explorer.Epoch {
		return g.AddCounterByKeyedForEffect(controller, explorer.ID, CounterPlusOne, 1, ask)
	}
	return ask.Run(g, nil)
}

// exploreAskTail is the half of an explore after its counter: ask whether
// the revealed card goes into the graveyard. Carry is the explorer, then the
// revealed card. An on-disk key: never renamed, never reused.
var exploreAskTail = RegisterTail("engine/explore-ask", func(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) < 2 {
		return nil
	}
	revealed := Card{InstanceID: o.Tail.Carry[1]}
	if p := g.playerByIDLocked(o.Tail.Player); p != nil {
		if c, ok := g.cardInZoneLocked(p.Library, revealed.InstanceID); ok {
			revealed = c
		}
	}
	return g.queueExploreGraveyardQuestionLocked(o.Tail.Source, o.Tail.Player, o.Tail.Carry[0], revealed, exploreRestFromTail(o.Tail.N, o.Tail.Word))
})

// exploreDoneTail is CR 701.44b's "has explored" after a move of the
// revealed card (to hand as a land, or to the graveyard on a yes), as a
// registered tail (#4361 S11b), so a CR 616 prompt on that move is a
// restore point. Source is the card whose effect told it to explore,
// Player the controller and Carry the explorer. An on-disk key: never
// renamed, never reused.
var exploreDoneTail = RegisterTail("engine/explore-done", func(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	revealed := uuid.Nil
	if len(o.Tail.Carry) > 1 {
		revealed = o.Tail.Carry[1]
	}
	return g.finishExploreLocked(o.Tail.Source, o.Tail.Carry[0], o.Tail.Player, revealed, exploreRestFromTail(o.Tail.N, o.Tail.Word))
})

// exploreDone is exploreDoneTail for one explore.
func exploreDone(source, explorer, controller, revealed uuid.UUID, rest exploreRest) KeyedTail {
	return KeyedTail{Ref: exploreDoneTail, Source: source, Player: controller, Carry: []uuid.UUID{explorer, revealed}, N: rest.More, Word: rest.tailWord()}
}

// finishExploreLocked is CR 701.44b: the permanent has explored.
//
// Caller must hold g.mu.
func (g *Game) finishExploreLocked(source, explorer, controller, revealed uuid.UUID, rest exploreRest) error {
	g.EmitEvent(Event{
		Kind:            EventExplored,
		Actor:           controller,
		CardID:          explorer,
		Source:          source,
		ExploreRevealed: revealed,
	})
	if rest.Again <= 0 && rest.More <= 0 {
		if len(rest.Queue) > 0 {
			// CR 701.44d: the next of the permanents told to explore.
			return exploreEachNext(g, source, controller, rest.Queue)
		}
		if ref, ok := TailRefByKey(rest.Then); ok {
			return KeyedTail{Ref: ref, Source: source, Player: controller, Carry: []uuid.UUID{explorer}}.Run(g, nil)
		}
		return nil
	}
	ref := ObjectRef{ID: explorer}
	if c := findBattlefieldCard(g, explorer); c != nil {
		ref.Epoch = c.ObjectEpoch
	}
	if rest.Again > 0 {
		// "Then it explores again": an explore the replacement made, which
		// no replacement applies to a second time (CR 614.5).
		rest.Again--
		return exploreAgain(g, source, ref, controller, rest, true)
	}
	// "Explores X times": the next explore, the same permanent again.
	rest.More--
	return exploreAgain(g, source, ref, controller, rest, false)
}

// exploreAgain is exploreLocked, reached through a variable so the package's
// initialisation order has no cycle (an explore's tails finish an explore).
// `replaced` is an explore a replacement already made (exploreNowLocked).
var exploreAgain func(g *Game, source uuid.UUID, explorer ObjectRef, controller uuid.UUID, rest exploreRest, replaced bool) error

func init() {
	exploreAgain = func(g *Game, source uuid.UUID, explorer ObjectRef, controller uuid.UUID, rest exploreRest, replaced bool) error {
		if replaced {
			return g.exploreNowLocked(source, explorer, controller, rest)
		}
		return g.exploreLocked(source, explorer, controller, rest)
	}
}

// queueExploreGraveyardQuestionLocked asks CR 701.44a's last clause:
// "may put the revealed card into their graveyard". The card is checked
// again when the answer arrives, because the library can change while
// the question is open; one that is no longer on top stays where it is.
//
// The answer is a registered key (#4298), so a table waiting on it is a
// restore point: the revealed card and the explorer ride the prompt as
// its carried IDs.
//
// Caller must hold g.mu.
func (g *Game) queueExploreGraveyardQuestionLocked(source, controller, explorer uuid.UUID, revealed Card, rest exploreRest) error {
	_, err := g.QueueConfirmKeyedForEffect(ConfirmPrompt{
		Chooser:      controller,
		Source:       source,
		Question:     "Explore — put " + nameOr(revealed.Name, "the revealed card") + " into your graveyard?",
		AcceptLabel:  "Put it into your graveyard",
		DeclineLabel: "Leave it on top",
		Word:         rest.word(),
	}, exploreGraveyardAnswer, []uuid.UUID{revealed.InstanceID, explorer})
	return err
}

// exploreGraveyardAnswer is the answer to the explore question. Carry is
// the revealed card, then the explorer.
var exploreGraveyardAnswer = RegisterConfirmThen("confirm/explore-graveyard", func(g *Game, r ConfirmAnswer) error {
	if len(r.Carry) < 2 {
		return nil
	}
	revealed, explorer, controller, source := r.Carry[0], r.Carry[1], r.Chooser, r.Source
	rest := exploreRestFromWord(r.Word)
	done := func(g *Game) error { return g.finishExploreLocked(source, explorer, controller, revealed, rest) }
	if !r.Accepted {
		return done(g)
	}
	p := g.playerByIDLocked(controller)
	if p == nil || p.Library.Size() == 0 || p.Library.Cards[p.Library.Size()-1].InstanceID != revealed {
		return done(g)
	}
	return g.routeCardKeyedLocked(zoneRoute{
		CardID:   revealed,
		Dst:      ZoneGraveyard,
		DstOwner: controller,
		Actor:    controller,
		Source:   source,
	}, exploreDone(source, explorer, controller, revealed, rest))
})

// exploreReason is the reveal banner: "<name> explores".
//
// Caller must hold g.mu.
func exploreReason(g *Game, explorer ObjectRef) string {
	name := "A permanent"
	if c := findBattlefieldCard(g, explorer.ID); c != nil {
		name = c.Effective().Name
	}
	return name + " explores"
}
