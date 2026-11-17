package game

import "github.com/google/uuid"

// dice_cost.go — dice rolled to pay an activation cost, and dice rolled
// as a permanent enters (#5682, CR 706).
//
// Both are rolls no resolving effect makes, so neither can hand "the
// rest of the effect" to RollDiceThenForEffect the way a card's effect
// does. Each still goes through that one door, so the roll is a
// replaceable event (Squid Fire Knight's sticker, Ichor Elixir), an
// ignored roll never happened (CR 706.6), and "whenever you roll one or
// more dice" sees the kept ones. What each hands the door is an engine
// tail that writes the kept results where the reader will look.
//
// # A roll as a cost
//
// Clay Golem: "{6}, Roll a d8: Monstrosity X, where X is the result."
//
//	CR 601.2h  "The player pays the total cost. First, they pay all
//	            costs that don't involve random elements or moving
//	            objects from the library to a public zone, in any order.
//	            Then they pay all remaining costs in any order."
//
// AbilityCost.RollDie is paid last, once the ability is on the stack
// (payRollDieCostLocked), and the result goes to that item's
// PaidCost.Rolled. A roll that a replacement makes ask "which roll do
// you ignore?" leaves that prompt open with the ability on the stack;
// the prompt blocks the table, so nobody acts before the result is
// known, and the answer writes it.
//
// # A roll as a permanent enters
//
// Neverwinter Hydra: "As this creature enters, roll X d6. It enters with
// a number of +1/+1 counters on it equal to the total of those results."
//
//	CR 614.1c   "Effects that read '[This permanent] enters with . . . ,'
//	             'As [this permanent] enters . . . ,' ... are replacement
//	             effects."
//	CR 614.12a  "If a replacement effect that modifies how a permanent
//	             enters the battlefield requires a choice, that choice is
//	             made before the permanent enters the battlefield."
//	CR 107.3m   the X of that replacement is the X of the spell that
//	             became the permanent; for any other entry it is 0.
//
// X is a fact about the cast, so the clause is an EntryCountersFromCast
// (RollXSides) and only a resolving permanent spell rolls: put onto the
// battlefield any other way, X is 0 and no die is rolled. The dice are
// rolled as the spell resolves, at the one place its entry is about to
// be built and before the CR 614 window opens (rollEntryDiceLocked), and
// kept on the item (StackItem.EntryRolls). The clause's counters are
// then seeded onto the entry event with every other cast-read clause,
// so Doubling Season and Hardened Scales see them. A roll that asks a
// question is one of the resolution's own prompts, and the entry waits
// for it with the rest of the resolution's tail (resolution_tail.go).
// A copy of the spell (CR 707.10) copies X and rolls its own dice as it
// resolves.

var (
	// rollCostTail writes a cost roll's kept results onto the ability
	// that paid it. An on-disk identity.
	rollCostTail = RegisterTail("dice/roll-cost", rollCostSettled)
	// entryRollTail writes an entry roll's kept results onto the
	// resolving spell. An on-disk identity.
	entryRollTail = RegisterTail("dice/entry-roll", entryRollSettled)
)

// payRollDieCostLocked pays "Roll a d<sides>" for the ability `itemID`,
// which is already on the stack. A no-op for a cost without one.
//
// Caller must hold g.mu (write).
func (g *Game) payRollDieCostLocked(player, source, itemID uuid.UUID, sides int) error {
	if sides <= 0 {
		return nil
	}
	return g.RollDiceThenForEffect(DiceRoll{
		Player: player, Source: source, Sides: sides, N: 1,
		Then: KeyedTail{Ref: rollCostTail, Player: player, Source: source, Carry: []uuid.UUID{itemID}},
	})
}

// rollCostSettled records the kept results on the ability, which is
// still on the stack: the prompt a paused roll waits on blocks the
// table.
func rollCostSettled(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	if item := g.StackMeta[o.Tail.Carry[0]]; item != nil {
		item.Paid.Rolled = append([]int(nil), o.Rolls...)
	}
	return nil
}

// entryDiceSidesFor is the kind of die the card's own "as this enters,
// roll X dN" clause rolls, or 0 when it declares none.
func entryDiceSidesFor(card Card) int {
	for _, clause := range EntersWithCountersFromCastFor(CatalogKey(card)) {
		if clause.RollXSides > 0 {
			return clause.RollXSides
		}
	}
	return 0
}

// rollEntryDiceLocked rolls the dice of a resolving permanent spell's
// own "as this enters, roll X dN" clause and keeps them on the item.
// Nothing is rolled for a spell without the clause, for X of 0 or less,
// for a spell cast face down (CR 708.2a: it has no such text), and for
// one that has already left the stack.
//
// Caller must hold g.mu (write), inside the spell's resolution.
func (g *Game) rollEntryDiceLocked(top Card, item *StackItem) {
	if item == nil || item.FaceDown != FaceDownNone || item.XValue <= 0 || !top.IsPermanent() {
		return
	}
	sides := entryDiceSidesFor(top)
	if sides <= 0 || g.spellMovedItselfLocked(top.InstanceID) {
		return
	}
	item.EntryRolls = nil
	if err := g.RollDiceThenForEffect(DiceRoll{
		Player: item.Controller, Source: top.InstanceID, Sides: sides, N: item.XValue,
		Then: KeyedTail{Ref: entryRollTail, Player: item.Controller, Source: top.InstanceID},
	}); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, Source: top.InstanceID, ErrorMsg: err.Error()})
	}
}

// entryRollSettled records the kept results on the resolving spell. The
// item is no longer in StackMeta: it is the resolving slot's, and once
// the resolution waits on the roll's own prompt it is the parked tail's
// too. After a restore those are two copies, so both are written.
func entryRollSettled(g *Game, o TailOutcome) error {
	id := o.Tail.Source
	if r := g.resolving; r != nil && r.hasCard && r.item != nil && r.item.ID == id {
		r.item.EntryRolls = append([]int(nil), o.Rolls...)
	}
	if t := g.pausedResolutionTail; t != nil && t.item != nil && t.item.ID == id {
		t.item.EntryRolls = append([]int(nil), o.Rolls...)
	}
	return nil
}
