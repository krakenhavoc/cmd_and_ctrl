package game

import "github.com/google/uuid"

// battle.go is the S27 battle lifecycle (CR 310): the defense
// counters a battle enters with, the protector its controller
// chooses, and the defeated trigger that fires as the last defense
// counter comes off.
//
// The shape mirrors sagas.go and the planeswalker starting-loyalty
// stamp deliberately: all three are printed rules that key off the
// card's TYPE rather than off a catalog entry, so a battle nobody has
// written a Spec for still enters with its printed defense and still
// gets a protector. What the catalog adds is the battle's ABILITIES.
//
// The one thing about battles that has no analogue elsewhere in the
// engine is the protector. Every other permanent is defended by its
// controller; a battle is cast by one player and defended by ANOTHER,
// chosen as it enters (CR 310.5). That inversion is the whole
// mechanic — "you and others can attack it" on a Siege means everyone
// except the protector — and it is why ProtectorPlayerID cannot be
// derived and has to be stored and chosen.

// PendingChoiceChooseProtector is the "as this battle enters, choose
// an opponent to protect it" prompt (CR 310.5). Answered with the
// same `{target: {kind: "player", id}}` payload the pick_target
// prompts use, because it is the same question shape — pick one
// player from a server-computed set — and reusing it means the
// client's existing player-highlight flow needs a route rather than a
// component.
//
// Declared here rather than in pending_choice.go's const block for
// the reason PendingChoiceEntryPayLife is: the kind, its queue path
// and its resume path are one mechanism, and keeping them in one file
// is what makes it reviewable.
//
// The chooser is the battle's CONTROLLER — the player who cast it —
// and the options are their opponents. It is not a target: a battle
// entering is not a spell or ability choosing targets (CR 115.1), so
// hexproof-style restrictions are irrelevant and there is nothing to
// re-check at resolution.
const PendingChoiceChooseProtector PendingChoiceKind = "choose_protector"

// CatalogBattleDefense returns the catalog's declared defense for an
// oracle ID, or 0. A FALLBACK only, exactly like
// CatalogStartingLoyalty: the printed value on Card.StartingDefense,
// stamped by the deck importer from Scryfall, always wins. The
// catalog answer exists for cards that never went through deck import
// — tokens, test fixtures, the demo seed.
var CatalogBattleDefense func(oracleID string) int

// stampBattleEntryLocked runs the two things that happen as a battle
// crosses onto the battlefield (CR 310.4, 310.5): the printed defense
// counters, and the protector choice.
//
// Called from fireETBHookLocked beside the starting-loyalty stamp and
// the Saga's lore counter, and unconditionally for the same reason —
// it is printed rules behaviour keyed on the card's type, so it must
// not sit behind the oracle-ID / nil-catalog guards.
//
// Caller must hold g.mu.
func (g *Game) stampBattleEntryLocked(cardID uuid.UUID, oracleID string) {
	card := findBattlefieldCard(g, cardID)
	if card == nil || !card.IsBattle() {
		return
	}
	controller := card.Controller
	// The already-has-counters guard makes the call idempotent for the
	// entry paths that fire the ETB hook more than once, and stops a
	// battle that entered with counters from some other effect being
	// double-stamped.
	if card.Counters[CounterDefense] <= 0 {
		defense := card.StartingDefense
		if defense <= 0 && CatalogBattleDefense != nil && oracleID != "" {
			defense = CatalogBattleDefense(oracleID)
		}
		if defense > 0 {
			// Through the CR 614 pipeline, like the loyalty stamp:
			// "enters with N defense counters" is a counter placement
			// an effect causes, so a Doubling Season doubles it.
			if err := g.AddCounterForEffect(cardID, CounterDefense, defense); err != nil {
				g.EmitEvent(Event{Kind: EventEffectError, Source: cardID, ErrorMsg: err.Error()})
			}
		}
		// A battle with nothing printed and nothing in the catalog
		// enters at zero and the CR 704.5p SBA takes it, exactly as a
		// planeswalker with no loyalty is taken by 704.5i. Inventing a
		// number here would make battles unbeatable.
	}
	g.queueChooseProtectorLocked(cardID, controller)
}

// queueChooseProtectorLocked asks the battle's controller which
// opponent protects it (CR 310.5).
//
// No prompt is queued when there is nobody to choose — a two-player
// game where the only opponent has been eliminated, or a fixture with
// one seat. In that case the battle simply has no protector, and
// defendingPlayerForAttackLocked falls back to its controller, which
// makes it un-attackable by the only player left. That is the weaker
// outcome and the safe one: the alternative, auto-assigning a seat,
// would silently pick a defender the controller never chose.
//
// A single eligible opponent is still PROMPTED rather than
// auto-assigned. Auto-assignment would be right on the rules and
// wrong on the interface: a prompt that only ever has one answer
// still tells the table who is defending, and the click is the
// cheapest way to make sure they saw it.
//
// Caller must hold g.mu.
func (g *Game) queueChooseProtectorLocked(battleID, controller uuid.UUID) {
	var options []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		options = append(options, p.ID)
	}
	if len(options) == 0 {
		return
	}
	name := "Battle"
	if c, ok := g.LookupCardForEffect(battleID); ok && c.Name != "" {
		name = c.Name
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:              PendingChoiceChooseProtector,
		Chooser:           controller,
		Count:             1,
		Source:            battleID,
		Reason:            name + " — choose an opponent to protect it",
		PickTargetPlayers: options,
		PickTargetMin:     1,
		PickTargetMax:     1,
	})
}

// ResolveChooseProtector records the controller's answer to a
// PendingChoiceChooseProtector prompt.
//
// Validated against the prompt's own option list rather than
// recomputed: a seat eliminated between the prompt and the answer is
// rejected here, and re-deriving "the controller's opponents" would
// have to agree with the list the client was shown.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveChooseProtector(choiceID, chooserID, protectorID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceChooseProtector {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	legal := false
	for _, id := range choice.PickTargetPlayers {
		if id == protectorID {
			legal = true
			break
		}
	}
	if !legal {
		return ErrIllegalTarget
	}
	if p := g.playerByIDLocked(protectorID); p == nil || p.Eliminated {
		return ErrIllegalTarget
	}
	battleID := choice.Source
	g.dequeueChoiceLocked(idx)
	if card := findBattlefieldCard(g, battleID); card != nil && card.IsBattle() {
		card.ProtectorPlayerID = protectorID
	}
	// The battle may already have been destroyed while the prompt sat
	// unanswered — a wrath in response to its own ETB. Dropping the
	// assignment silently is right; the prompt is drained either way
	// so priority can move again.
	g.runStateChecksLocked()
	return nil
}

// emitBattleDefeatedLocked announces a battle's defeat so a
// DefeatedTrigger can harvest it (CR 310.9).
//
// Emitted from the state-based-action pass IMMEDIATELY BEFORE the
// CR 704.5p move that puts the battle in the graveyard, which is what
// makes the trigger's source findable on the battlefield when the
// harvester walks it. Doing it after the move would send the harvest
// down the LTB path, where the only characteristics available are the
// last-known-information snapshot — correct for a dies trigger and
// unnecessarily lossy for this one.
//
// Caller must hold g.mu.
func (g *Game) emitBattleDefeatedLocked(battleID uuid.UUID) {
	c := findBattlefieldCard(g, battleID)
	if c == nil {
		return
	}
	g.EmitEvent(Event{
		Kind:   EventBattleDefeated,
		Actor:  c.Controller,
		Source: battleID,
		Target: battleID,
		CardID: battleID,
	})
}
