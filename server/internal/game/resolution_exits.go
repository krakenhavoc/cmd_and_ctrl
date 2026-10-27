package game

import "github.com/google/uuid"

// resolution_exits.go — the replacements of "put it into its owner's
// graveyard as it resolves" (CR 608.2n), and CR 616.1's choice between
// them (ADR 0107 §3, #1854).
//
//	CR 616.1  "If two or more replacement and/or prevention effects are
//	           attempting to modify the way an event affects an object
//	           or player, the affected object's controller (or its owner
//	           if it has no controller) or the affected player chooses
//	           one to apply …"
//	CR 616.1f "Once the chosen effect has been applied, this process is
//	           repeated (taking into account only replacement or
//	           prevention effects that would now be applicable) …"
//
// These replace that one event, and each is a static ability of the
// spell, a rule of the game or an effect from outside the spell, never
// a self-replacement effect of its resolution (CR 614.15), so none
// takes precedence under CR 616.1a:
//
//   - A GRANTED CAST'S "if that spell would be put into your graveyard,
//     exile it instead" (StackItem.ExileInsteadOfGraveyard, #3863:
//     Torrential Gearhulk, Emet-Selch of the Third Seat). Offered
//     first, so a prompt that ends unanswered takes the exile.
//   - REBOUND (CR 702.88a, rebound.go): exile it, and cast it free at
//     its controller's next upkeep. Printed, or given to the spell by
//     the stack step of the layer pass (spell_keywords.go).
//   - BUYBACK (CR 702.27a): put it into its owner's hand.
//   - THE ADVENTURE EXILE (CR 715.3d): exile it, and its controller may
//     cast the creature later (adventure.go).
//
// Each sends the card somewhere other than a graveyard, so once one is
// applied the others no longer apply (CR 616.1f): one choice, never
// a sequence. Flashback (CR 702.34a) is not in the list. It replaces
// every exit from the stack, so routeStackCardToGraveyardLocked
// applies it before asking anything, and a spell cast with flashback
// was cast from a graveyard and has no rebound to compete with.
//
// The choice is a PendingChoiceOptionPick asked of the spell's
// controller while the card waits on the stack, exactly as the CR 903.9
// commander prompt makes it wait: a pending choice stops priority, so
// nothing resolves past it. The continuation is keyed and holds IDs and
// words only, never a pointer into the game, so it resolves against
// whichever game the answer arrives in, a restored one included.

// resolutionExit is one applicable replacement: the words the chooser
// reads, which replacement it is (one of the resolutionExit* kinds) and
// the IDs it needs. It is plain data (#4361 slice 6), so the choice
// between two of them rides a keyed option pick and a table waiting on
// it is a restore point.
type resolutionExit struct {
	label string
	kind  string
	// card is the spell's card, owner its owner, and controller the
	// stack item's controller (uuid.Nil with no item).
	card, owner, controller uuid.UUID
	// counter is the counter kind of the exile-with-a-counter exit.
	counter string
}

// The exits' kinds, written on a waiting prompt (resolutionExitThen's
// Words): on-disk identities, never renamed, never reused.
const (
	resolutionExitExileInstead = "exile-instead"
	resolutionExitRebound      = "rebound"
	resolutionExitBuyback      = "buyback"
	resolutionExitWithCounter  = "exile-with-counter"
	resolutionExitAdventure    = "adventure"
)

// apply rewrites the route the way this exit replaces it, and returns
// what follows the move (zero for nothing) as a registered tail, which
// the caller routes the card with (routeCardKeyedLocked), so a CR 616
// prompt on the exit is a restore point (#4361 S11b).
func (e resolutionExit) apply(r *zoneRoute) KeyedTail {
	cardID, controller := e.card, e.controller
	switch e.kind {
	case resolutionExitExileInstead:
		r.Dst, r.DstOwner, r.Actor = ZoneExile, uuid.Nil, e.owner
	case resolutionExitRebound:
		r.Dst, r.DstOwner, r.Actor = ZoneExile, uuid.Nil, controller
		return KeyedTail{Ref: reboundExiledTail, Player: controller, Carry: []uuid.UUID{cardID}}
	case resolutionExitBuyback:
		r.Dst = ZoneHand
	case resolutionExitWithCounter:
		r.Dst, r.DstOwner, r.Actor = ZoneExile, uuid.Nil, controller
		return KeyedTail{Ref: exiledWithCounterTail, Player: controller, Carry: []uuid.UUID{cardID}, Word: e.counter}
	case resolutionExitAdventure:
		// CR 715.3d names the spell's CONTROLLER twice — "its controller
		// exiles it" and "that player may play it" — so a stolen
		// Adventure is the thief's to cast later (ADR 0104, owner
		// decision 4). The owner stands in only for an item with no
		// controller, which no cast produces.
		if controller == uuid.Nil {
			controller = e.owner
		}
		r.Dst, r.DstOwner, r.Actor = ZoneExile, uuid.Nil, controller
		return KeyedTail{Ref: adventureExiledTail, Player: controller, Carry: []uuid.UUID{cardID}}
	}
	return KeyedTail{}
}

// The exits' continuations as registered tails (#4361 S11b). Player is
// the stack item's controller (the Adventure's: its owner when it had
// none) and Carry the spell's card; the exile with a counter's Word is
// the counter kind. On-disk keys: never renamed, never reused.
var (
	// reboundExiledTail: CR 702.88a's delayed trigger, about the card IN
	// EXILE, created once it is there.
	reboundExiledTail = RegisterTail("engine/rebound-exiled", func(g *Game, o TailOutcome) error {
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		g.scheduleReboundLocked(o.Tail.Carry[0], o.Tail.Player)
		return nil
	})
	// exiledWithCounterTail: Goliath Daydreamer's counter, put on once
	// the card is really in exile (exile_with_counter.go).
	exiledWithCounterTail = RegisterTail("engine/exiled-with-counter", func(g *Game, o TailOutcome) error {
		if len(o.Tail.Carry) == 0 || exiledCardByIDLocked(g, o.Tail.Carry[0]) == nil {
			return nil
		}
		// #3942: the spell's controller puts it.
		return g.AddCounterByForEffect(o.Tail.Player, o.Tail.Carry[0], o.Tail.Word, 1)
	})
	// adventureExiledTail: CR 715.4's permission to cast the creature,
	// granted once the card is in exile (adventure.go).
	adventureExiledTail = RegisterTail("engine/adventure-exiled", func(g *Game, o TailOutcome) error {
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		g.grantAdventureCastFromExileLocked(o.Tail.Carry[0], o.Tail.Player)
		return nil
	})
)

// resolutionExitsLocked lists the replacements that apply to a spell
// that has just RESOLVED, in the order they are offered: the granted
// cast's exile, rebound, buyback, the exile with a counter, the
// Adventure exile. Empty for an ordinary spell, which goes to its
// owner's graveyard.
//
// Caller must hold g.mu.
func resolutionExitsLocked(c Card, item *StackItem) []resolutionExit {
	var out []resolutionExit
	exit := func(label, kind string) resolutionExit {
		e := resolutionExit{label: label, kind: kind, card: c.InstanceID, owner: c.Owner}
		if item != nil {
			e.controller = item.Controller
		}
		return e
	}
	if exilesInsteadOfGraveyard(item) {
		// #3863: the granted cast's replacement. Offered first: of the
		// exits, it is the one that gives the player nothing back.
		out = append(out, exit("Exile it instead of putting it into your graveyard", resolutionExitExileInstead))
	}
	if spellRebounds(c, item) {
		// CR 702.88a: exile it and, at the beginning of its
		// controller's next upkeep, offer the free cast. The delayed
		// trigger rides the route's continuation for the reason the
		// Adventure grant does: it is about the card IN EXILE, and is
		// created only once the card is there.
		out = append(out, exit("Rebound — exile it, and you may cast it free at your next upkeep", resolutionExitRebound))
	}
	if item != nil && OptionalCostTimesPaid(c, item.Paid.OptionalCosts, BuybackKey) > 0 {
		// CR 702.27a. Through the SAME exit primitive, so a bought-back
		// commander still gets its CR 903.9 choice and a replacement
		// watching the stack exit still sees one.
		out = append(out, exit("Buyback — return it to its owner's hand", resolutionExitBuyback))
	}
	if item != nil && !item.IsCopy && item.ExileWithCounter != "" {
		// Goliath Daydreamer: exile it with a counter on it. The counter
		// goes on once the card is really in exile (exile_with_counter.go).
		e := exit("Exile it with a "+item.ExileWithCounter+" counter on it", resolutionExitWithCounter)
		e.counter = item.ExileWithCounter
		out = append(out, e)
	}
	if castAsAdventure(c) {
		// CR 715.3d, and CR 715.4's permission rides the route's
		// continuation rather than the next line: a route that paused
		// on a CR 903.9 prompt finishes later, and the grant has to land
		// when it does. See adventure.go.
		out = append(out, exit("Adventure — exile it, and you may cast the creature later", resolutionExitAdventure))
	}
	return out
}

// chooseResolutionExitLocked asks the spell's controller which of
// `exits` applies (CR 616.1), and routes the card once they answer.
// The route is the one routeStackCardToGraveyardLocked builds, to the
// owner's graveyard by a rule (resolutionExitRouteOf); the chosen exit
// rewrites it.
//
// A prompt that cannot be asked (the controller has left) or that ends
// unanswered takes the first exit in the offered order. That is the
// rest of the resolution running with the question unasked (the
// NoChoiceIndex rule), and the order is the one written above.
//
// Keyed (#4361 slice 6): the card, its owner, the item's controller, the
// counter kind and the exits' kinds ride the prompt as data
// (resolutionExitThen), so a table waiting on the answer is a restore
// point.
//
// Caller must hold g.mu (write).
func (g *Game) chooseResolutionExitLocked(c Card, item *StackItem, exits []resolutionExit) error {
	chooser := c.Owner
	if item != nil && item.Controller != uuid.Nil {
		chooser = item.Controller
	}
	options := make([]ChoiceOption, len(exits))
	words := make([]string, 1, len(exits)+1)
	for i, e := range exits {
		options[i] = ChoiceOption{Label: e.label}
		words = append(words, e.kind)
		if e.counter != "" {
			words[0] = e.counter
		}
	}
	carry := []uuid.UUID{exits[0].card, exits[0].owner, exits[0].controller}
	id := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  chooser,
		Source:   c.InstanceID,
		Question: c.Name + " is resolving — choose which replacement applies",
		Options:  options,
		ThenKey:  resolutionExitThen,
		Carry:    carry,
		Words:    words,
	})
	if id == uuid.Nil {
		return routeResolutionExitLocked(g, carry, words, NoChoiceIndex)
	}
	return nil
}

// resolutionExitRouteOf is the route a resolved spell's card takes to
// its owner's graveyard before any exit rewrites it (CR 608.2n; #1320:
// no spell or ability is moving it, a rule is).
func resolutionExitRouteOf(card, owner uuid.UUID) zoneRoute {
	return zoneRoute{
		CardID: card, Dst: ZoneGraveyard, DstOwner: owner,
		Cause: MoveCause{Kind: MoveCauseRule},
	}
}

// resolutionExitThen is the answer to CR 616.1's choice between a
// resolved spell's exits. Carry is the card, its owner and the stack
// item's controller; Words[0] is the exile-with-a-counter exit's counter
// kind ("" without one) and Words[1:] the exits' kinds in offered order.
// None costs mana, so the index is an offset into them. An on-disk key:
// never renamed, never reused. Registered from init: the route it runs
// can resolve further spells' exits, which name this key.
var resolutionExitThen OptionPickThen

func init() {
	resolutionExitThen = RegisterOptionPickThen("option-pick/resolution-exit", func(g *Game, r OptionPicked) error {
		return routeResolutionExitLocked(g, r.Carry, r.Words, r.Index)
	})
}

// routeResolutionExitLocked routes the card by the exit at `index`, or
// by the first one when the index is out of range (NoChoiceIndex).
//
// Caller must hold g.mu (write).
func routeResolutionExitLocked(g *Game, carry []uuid.UUID, words []string, index int) error {
	if len(carry) < 3 || len(words) < 2 {
		return nil
	}
	kinds := words[1:]
	if index < 0 || index >= len(kinds) {
		index = 0
	}
	e := resolutionExit{kind: kinds[index], card: carry[0], owner: carry[1], controller: carry[2]}
	if e.kind == resolutionExitWithCounter {
		e.counter = words[0]
	}
	r := resolutionExitRouteOf(e.card, e.owner)
	then := e.apply(&r)
	return g.routeCardKeyedLocked(r, then)
}

// exilesInsteadOfGraveyard reports whether the spell carries a granted
// cast's "if that spell would be put into your graveyard, exile it
// instead" (StackItem.ExileInsteadOfGraveyard, #3863). Never for a
// copy, which ceases to exist instead (CR 707.10a).
func exilesInsteadOfGraveyard(item *StackItem) bool {
	return item != nil && !item.IsCopy && item.ExileInsteadOfGraveyard
}
