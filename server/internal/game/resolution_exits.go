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
// Three rules replace that one event, and each is a static ability of
// the spell or a rule of the game, never a self-replacement effect of
// its resolution (CR 614.15), so none takes precedence under CR 616.1a:
//
//   - REBOUND (CR 702.88a, rebound.go): exile it, and cast it free at
//     its controller's next upkeep. Printed, or given to the spell by
//     the stack step of the layer pass (spell_keywords.go).
//   - BUYBACK (CR 702.27a): put it into its owner's hand.
//   - THE ADVENTURE EXILE (CR 715.3d): exile it, and its controller may
//     cast the creature later (adventure.go).
//
// Each sends the card somewhere other than a graveyard, so once one is
// applied the other two no longer apply (CR 616.1f): one choice, never
// a sequence. Flashback (CR 702.34a) is not in the list. It replaces
// every exit from the stack, so routeStackCardToGraveyardLocked
// applies it before asking anything, and a spell cast with flashback
// was cast from a graveyard and has no rebound to compete with.
//
// The choice is a PendingChoiceOptionPick asked of the spell's
// controller while the card waits on the stack, exactly as the CR 903.9
// commander prompt makes it wait: a pending choice stops priority, so
// nothing resolves past it. The continuation holds IDs only, never a
// pointer into the game, so it resolves against whichever game the
// answer arrives in.

// resolutionExit is one applicable replacement: the words the chooser
// reads and what it does to the route.
type resolutionExit struct {
	label string
	apply func(r *zoneRoute)
}

// resolutionExitsLocked lists the replacements that apply to a spell
// that has just RESOLVED, in the order they are offered: rebound,
// buyback, the Adventure exile. Empty for an ordinary spell, which goes
// to its owner's graveyard.
//
// Caller must hold g.mu.
func resolutionExitsLocked(c Card, item *StackItem) []resolutionExit {
	var out []resolutionExit
	if spellRebounds(c, item) {
		// CR 702.88a: exile it and, at the beginning of its
		// controller's next upkeep, offer the free cast. The delayed
		// trigger rides the continuation for the reason the Adventure
		// grant does: it is about the card IN EXILE, and is created only
		// once the card is there.
		cardID, controller := c.InstanceID, item.Controller
		out = append(out, resolutionExit{
			label: "Rebound — exile it, and you may cast it free at your next upkeep",
			apply: func(r *zoneRoute) {
				r.Dst, r.DstOwner, r.Actor = ZoneExile, uuid.Nil, controller
				r.then = func(g *Game) error {
					g.scheduleReboundLocked(cardID, controller)
					return nil
				}
			},
		})
	}
	if item != nil && OptionalCostTimesPaid(c, item.Paid.OptionalCosts, BuybackKey) > 0 {
		// CR 702.27a. Through the SAME exit primitive, so a bought-back
		// commander still gets its CR 903.9 choice and a replacement
		// watching the stack exit still sees one.
		out = append(out, resolutionExit{
			label: "Buyback — return it to its owner's hand",
			apply: func(r *zoneRoute) { r.Dst = ZoneHand },
		})
	}
	if castAsAdventure(c) {
		// CR 715.3d, and CR 715.4's permission rides the route's
		// continuation rather than the next line: a route that paused
		// on a CR 903.9 prompt finishes later, and the grant has to land
		// when it does. See adventure.go.
		//
		// CR 715.3d names the spell's CONTROLLER twice — "its controller
		// exiles it" and "that player may play it" — so a stolen
		// Adventure is the thief's to cast later (ADR 0104, owner
		// decision 4). The owner stands in only for an item with no
		// controller, which no cast produces.
		cardID := c.InstanceID
		controller := c.Owner
		if item != nil && item.Controller != uuid.Nil {
			controller = item.Controller
		}
		out = append(out, resolutionExit{
			label: "Adventure — exile it, and you may cast the creature later",
			apply: func(r *zoneRoute) {
				r.Dst, r.DstOwner, r.Actor = ZoneExile, uuid.Nil, controller
				r.then = func(g *Game) error {
					g.grantAdventureCastFromExileLocked(cardID, controller)
					return nil
				}
			},
		})
	}
	return out
}

// chooseResolutionExitLocked asks the spell's controller which of
// `exits` applies (CR 616.1), and routes the card once they answer.
// `r` is the route as the caller built it, to the owner's graveyard;
// the chosen exit rewrites it.
//
// A prompt that cannot be asked (the controller has left) or that ends
// unanswered takes the first exit in the offered order. That is the
// rest of the resolution running with the question unasked (the
// NoChoiceIndex rule), and the order is the one written above.
//
// Caller must hold g.mu (write).
func (g *Game) chooseResolutionExitLocked(c Card, item *StackItem, r zoneRoute, exits []resolutionExit) error {
	chooser := c.Owner
	if item != nil && item.Controller != uuid.Nil {
		chooser = item.Controller
	}
	route := func(g *Game, index int) error {
		if index < 0 || index >= len(exits) {
			index = 0
		}
		chosen := r
		exits[index].apply(&chosen)
		_, err := g.routeCardToZoneLocked(chosen)
		return err
	}
	options := make([]ChoiceOption, len(exits))
	for i, e := range exits {
		options[i] = ChoiceOption{Label: e.label}
	}
	id := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  chooser,
		Source:   c.InstanceID,
		Question: c.Name + " is resolving — choose which replacement applies",
		Options:  options,
		Then:     route,
	})
	if id == uuid.Nil {
		return route(g, NoChoiceIndex)
	}
	return nil
}
