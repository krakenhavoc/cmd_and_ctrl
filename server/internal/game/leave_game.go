package game

import "github.com/google/uuid"

// leave_game.go is CR 800.4a: what happens to a departed player's
// OBJECTS when they concede or lose. The rest of the departure — the
// seat flag, the stack cleanup, the game-over check and the turn
// rotation — lives in mutations.go and rotation.go; this file is only
// the part the rule spells out object by object.
//
// The rule, in its order (CR 800.4a, Aug 7 2026):
//
//  1. every object the departing player OWNS leaves the game;
//  2. effects that gave that player control of anything END;
//  3. non-card objects they controlled on the stack cease to exist;
//  4. anything STILL controlled by them — another player's permanent
//     they had taken by an effect that did not end at step 2 — is
//     exiled.
//
// It is explicitly NOT a state-based action: it happens the moment the
// player leaves, not at the next priority boundary. So it runs inside
// leaveGameLocked, which is the one door both Concede and the SBA loss
// loop go through.
//
// Step 3 is older than this file (`cleanupStackForEliminatedLocked`,
// S13.1) and stays where it is; leaveGameObjectsLocked runs after it,
// which is also what folds the spell CARDS that step used to exile
// into step 1's removal — a card owned by the player who left leaves
// the game, it does not sit in exile.
//
// ADR 0060 has the decisions: why removal is a real removal rather
// than a "left the game" bucket, why nothing here fires a leaves-the-
// battlefield or dies trigger, and what happens to their commander.

// survivingSeatsLocked counts the seats still in the game. Caller must
// hold g.mu.
func (g *Game) survivingSeatsLocked() int {
	n := 0
	for _, s := range g.Seats {
		if !s.Eliminated {
			n++
		}
	}
	return n
}

// leaveGameObjectsLocked performs steps 1, 2 and 4 of CR 800.4a for
// the player who has just left, in that order — the order matters,
// and is the difference between a Mind Control that leaves with its
// controller (the creature goes back to its owner) and one that does
// not (the creature would be exiled).
//
// Caller must hold g.mu, and must already have set p.Eliminated.
func (g *Game) leaveGameObjectsLocked(playerID uuid.UUID) {
	// 0. CR 702.143f, ADR 0069 decision 5: when a player leaves the
	//    game, all cards they own that are face down in exile are
	//    revealed. BEFORE step 1, because step 1 takes those cards out
	//    of the game and there is nothing left to reveal afterwards.
	//    A reveal is not a zone change, so this does not disturb the
	//    no-zone-move posture the sweep below deliberately takes.
	g.revealFaceDownOwnedByLocked(playerID)

	// 1. Everything they own leaves the game.
	removed := g.removeObjectsOwnedByLocked(playerID)

	// 2. Their control effects end. Every control-changing effect the
	//    engine has is a layer-2 continuous effect sourced from a
	//    permanent (ADR 0012 / the S24 Mind Control work), so step 1
	//    has already taken the SOURCE away and all that is owed is a
	//    recompute: `printedCharacteristic` seeds layer 2 from
	//    `BaseController`, so the creature goes back to whoever
	//    controlled it when it entered. The bump is by hand because
	//    nothing emitted an event for the removal above — deliberately,
	//    see removeObjectsOwnedByLocked — and the recompute is
	//    unconditional because step 4 reads Card.Controller off the
	//    back of it.
	g.layerVersion.Add(1)
	g.recomputeLayersLocked()

	// 4. Anything still under their control belongs to somebody else
	//    and has nobody to control it. Exiled.
	g.exileStillControlledLocked(playerID)

	// CR 800.4b/d, and the same reasoning cleanupStackForEliminatedLocked
	// gives for PendingTriggers: nothing goes on the stack under the
	// control of a player who is not in the game. A delayed trigger
	// they scheduled has no controller to put it on the stack for, and
	// the elimination event itself may have harvested a trigger of
	// theirs a moment ago.
	g.dropDelayedTriggersForLocked(playerID)
	g.dropPendingTriggersForLocked(playerID)

	if removed == 0 {
		return
	}
	// Two prompt families can be left asking about objects that are no
	// longer there. Both are the same re-checks executeBattlefieldLeaveLocked
	// runs after any other permanent leaves the battlefield, for the
	// same reason: a queued choice stops priority from passing, so the
	// state-check loop is exactly what does NOT run while one is open.
	g.pruneSacrificeChoicesLocked()
	g.pruneStaleZoneChangeChoicesLocked()
}

// removeObjectsOwnedByLocked drops every card owned by the departed
// player out of every zone the game tracks, and reports how many went.
//
// It removes them from the zone slices directly instead of routing
// them through MoveCard, and that is the whole point: leaving the game
// is not a zone change. No EventZoneMove, no EventLTB, no EventETB, so
// no dies trigger, no leaves-the-battlefield trigger, and no CR 603.10
// last-known-information snapshot for an event that never happened.
// The departing player's Blood Artist does not see their own board go.
//
// Every zone, not just the departed player's own four: a card they own
// can be on the shared battlefield under another player's control
// (stolen), in shared exile, on the stack, and — through a sandbox
// move or an effect that puts a card into another player's zone — in
// somebody else's hand or graveyard. Ownership is the test the rule
// names, so ownership is the test here.
//
// Caller must hold g.mu.
func (g *Game) removeObjectsOwnedByLocked(playerID uuid.UUID) int {
	removed := 0
	sweep := func(z *Zone) {
		if z == nil || len(z.Cards) == 0 {
			return
		}
		kept := make([]Card, 0, len(z.Cards))
		for _, c := range z.Cards {
			if c.Owner == playerID {
				removed++
				g.forgetObjectLocked(c.InstanceID)
				continue
			}
			kept = append(kept, c)
		}
		if len(kept) == len(z.Cards) {
			return
		}
		if len(kept) == 0 {
			z.Cards = nil
			return
		}
		z.Cards = kept
	}
	sweep(g.Battlefield)
	sweep(g.Stack)
	sweep(g.Exile)
	for _, p := range g.Seats {
		sweep(p.Library)
		sweep(p.Hand)
		sweep(p.Graveyard)
		sweep(p.Command)
		// #623 / CR 114: emblems are owned objects in the command
		// zone, so they leave with their owner like everything else.
		// This is the ONLY thing in the game that removes one.
		sweep(p.Emblems)
	}
	if removed > 0 {
		g.recomputeSplitSecondLocked()
	}
	return removed
}

// forgetObjectLocked drops the per-object bookkeeping that outlives a
// card's zone — the maps keyed by instance ID. An ordinary zone change
// keeps all of it (a bounced planeswalker that comes back this turn
// has still activated a loyalty ability this turn); an object leaving
// the game keeps none of it, because there is no object left to
// remember anything about.
//
// Caller must hold g.mu.
func (g *Game) forgetObjectLocked(cardID uuid.UUID) {
	if g.StackMeta != nil {
		delete(g.StackMeta, cardID)
	}
	if g.LoyaltyActivatedThisTurn != nil {
		delete(g.LoyaltyActivatedThisTurn, cardID)
	}
	if g.lastKnownBattlefield != nil {
		delete(g.lastKnownBattlefield, cardID)
	}
	if g.lastKnownTriggerIdentity != nil {
		delete(g.lastKnownTriggerIdentity, cardID)
	}
}

// exileStillControlledLocked is CR 800.4a's last clause: once the
// departed player's own objects are gone and their control effects
// have ended, anything on the battlefield that is STILL controlled by
// them is exiled.
//
// What reaches this today is a permanent whose control baseline IS the
// departed player without a continuous effect saying so — another
// player's card that entered the battlefield under the departed
// player's control (reanimated out of an opponent's graveyard, put
// onto the battlefield off the top of their library). A Mind Control
// does NOT reach it: the Aura left with its owner one step ago, so the
// recompute has already handed the creature back (CR 613.1b), which is
// the printed behaviour and the reason the two steps are ordered.
//
// This one IS a zone change, so it goes through the ordinary
// battlefield exit and its triggers fire — the creature really is
// exiled from the battlefield, and a survivor's "whenever a creature
// leaves the battlefield" watcher saw it happen.
//
// Caller must hold g.mu.
func (g *Game) exileStillControlledLocked(playerID uuid.UUID) {
	if g.Battlefield == nil {
		return
	}
	var stuck []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == playerID {
			stuck = append(stuck, c.InstanceID)
		}
	}
	for _, id := range stuck {
		if err := g.executeBattlefieldLeaveLocked(id, ZoneExile, uuid.Nil, nil); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "leaving the game: exile of a still-controlled permanent failed: " + err.Error(),
			})
		}
	}
}

// exileGhostControlledLocked is CR 800.4c: a control-changing effect
// ending hands its permanent back to the player who controlled it
// before, and THAT player may have left the game in the meantime. The
// permanent is exiled rather than handed to a ghost.
//
// It is the same outcome CR 800.4a gives a permanent still controlled
// by the player who has just left, one effect-ending later — the
// difference is only when the effect ends. The A-steals-it, B-departs,
// A's-effect-ends order is what reaches this and not
// exileStillControlledLocked: at the moment B left, A controlled the
// permanent, so there was nothing to exile.
//
// The rules perform this as part of the effect ending; the engine
// performs it on the next state-based-action pass, which is the first
// moment after an effect ends that anything looks at the whole board.
// Reports how many it exiled, so the SBA pass can count itself as
// having fired and run again.
//
// A game that has ENDED keeps its final board (ADR 0060 Decision 5)
// and no state check runs on it, so this never strips one.
//
// Caller must hold g.mu.
func (g *Game) exileGhostControlledLocked() int {
	if g.Battlefield == nil || g.State != StateActive {
		return 0
	}
	var ghosts []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		p := g.playerByIDLocked(c.Controller)
		if p == nil || !p.Eliminated {
			continue
		}
		ghosts = append(ghosts, c.InstanceID)
	}
	for _, id := range ghosts {
		if err := g.executeBattlefieldLeaveLocked(id, ZoneExile, uuid.Nil, nil); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "CR 800.4c: exile of a permanent left to a departed controller failed: " + err.Error(),
			})
		}
	}
	return len(ghosts)
}

// dropDelayedTriggersForLocked removes the departed player's delayed
// triggered abilities (CR 603.7) from the queue. See the CR 800.4b/d
// note at the call site.
//
// Caller must hold g.mu.
func (g *Game) dropDelayedTriggersForLocked(playerID uuid.UUID) {
	if len(g.DelayedTriggers) == 0 {
		return
	}
	kept := g.DelayedTriggers[:0]
	for _, d := range g.DelayedTriggers {
		if d == nil || d.Controller != playerID {
			kept = append(kept, d)
		}
	}
	g.DelayedTriggers = kept
	if len(g.DelayedTriggers) == 0 {
		g.DelayedTriggers = nil
	}
}

// dropPendingTriggersForLocked removes every triggered ability waiting
// in the APNAP queue whose controller has left the game. Extracted
// from cleanupStackForEliminatedLocked so the departure can re-run it
// after the elimination event and the CR 800.4a exiles have had their
// chance to harvest one more.
//
// Caller must hold g.mu.
func (g *Game) dropPendingTriggersForLocked(playerID uuid.UUID) {
	if len(g.PendingTriggers) == 0 {
		return
	}
	kept := g.PendingTriggers[:0]
	for _, t := range g.PendingTriggers {
		if t == nil || t.Controller != playerID {
			kept = append(kept, t)
		}
	}
	g.PendingTriggers = kept
}

// ---------------------------------------------------------------------
// CR 800.4g/h — a choice a departed player owed on somebody ELSE's
// object is reassigned, not dropped (#902).
// ---------------------------------------------------------------------
//
// The rules, verbatim (Aug 7 2026):
//
//	800.4f  If an object requires a player who has left the game to pay
//	        a cost or choose whether to pay a cost, that cost is not
//	        paid.
//	800.4g  If an object requires a player who has left the game to make
//	        a choice other than whether to pay a cost, the controller of
//	        the object chooses another player to make that choice. If the
//	        original choice was to be made by an opponent of the
//	        controller of the object, that player chooses another
//	        opponent if possible.
//	800.4h  If a rule requires a player who has left the game to make a
//	        choice, the next player in turn order makes that choice.
//
// Before #902 the engine dropped every prompt whose chooser had left
// (#864 / PR #868), which is right for their own material and wrong for
// a prompt a living player's card pointed at them. The drop is still
// the answer for most kinds and is still what happens when nobody is
// left to ask; what changes is that a prompt an object owes and a
// survivor can answer now changes hands.
//
// Three pieces, and no fourth:
//
//   - choiceReassignDecisions — one row per PendingChoiceKind, the
//     never-reassign table ADR 0060's amendment prints.
//   - choiceInheritorLocked — one inheritor policy: 800.4g's second
//     sentence, walked in turn order from the departed seat.
//     800.4h's "a rule requires" arm reassigns nothing today, because
//     every rule-required prompt the engine has is about the asked
//     player's own material and is `false` in the table.
//   - reassignDepartedChoiceLocked — the predicate the departure sweep
//     calls, which puts the two together and hands the winner to
//     reassignChoiceLocked (pending_choice.go) to perform.
//
// No kind is special-cased by card, and nothing here reads a card name.

// choiceReassignDecisions classifies every PendingChoiceKind against
// CR 800.4g: does the choice still mean something when somebody else
// makes it?
//
// DENY BY DEFAULT, and deliberately the opposite default from
// choiceGateDecisions (choice_gate.go). An unclassified kind is
// DROPPED, which is exactly the pre-#902 behaviour — safe, already
// shipped, and not a wedge. The failure mode of the wrong default here
// is a prompt that ends instead of moving, not a prompt handed to a
// seat it makes no sense for.
// TestEveryChoiceKindHasAReassignmentDecision fails until a new kind
// gets a row anyway.
//
// `true` is only for kinds that can be about ANOTHER player's object.
// Everything else is one of three shapes:
//
//   - CR 800.4f — the choice is a cost, or whether to pay one. The
//     rule says the cost is simply not paid, so there is nobody to ask.
//   - The material is the departed player's own, and CR 800.4a has
//     already taken it out of the game: their library, their hand,
//     their permanents, their mana pool, their loop.
//   - The prompt is welded to a pipeline that has its own terminal
//     outcome for a chooser who left — the CR 616 replacement pair,
//     whose drop settles the paused event through
//     finishDroppedReplacementLocked (#808).
var choiceReassignDecisions = map[PendingChoiceKind]bool{
	// --- CR 800.4g reassigns these -------------------------------
	//
	// pick_target: a triggered ability's CR 603.3d target pick. The
	// legal set is the whole board, and the trigger's source can be a
	// permanent a survivor controls.
	PendingChoicePickTarget: true,
	// trigger_prompt: a CR 603.5 "you may" whose chooser was pointed
	// at somebody other than the source's controller
	// (TriggerOptionalPrompt.Chooser).
	PendingChoiceTriggerPrompt: true,
	// choose_cards / discard_from_hand: Thoughtseize-shaped — the
	// chooser picks out of FromPlayer's pool, and when that is not
	// their own the question survives them.
	PendingChoiceChooseCards:     true,
	PendingChoiceDiscardFromHand: true,
	// option_pick: the second half of a pile split is a living
	// player's cards. Torment of Hailfire's "each opponent chooses"
	// is the same KIND and is still dropped — by the material test in
	// reassignDepartedChoiceLocked, not by this row.
	PendingChoiceOptionPick: true,

	// --- CR 800.4f: a cost, or whether to pay one ----------------
	PendingChoicePayUnless:    false,
	PendingChoiceEntryPayLife: false,
	// mana_pick is a cost's other half: the mana would enter a pool
	// that has left the game with its player.
	PendingChoiceMana: false,

	// --- their own material, already gone (CR 800.4a) ------------
	PendingChoiceSacrifice:     false,
	PendingChoiceLegendRule:    false,
	PendingChoiceScry:          false,
	PendingChoiceSurveil:       false,
	PendingChoiceLookAtTop:     false,
	PendingChoiceSearchLibrary: false,
	PendingChoiceMayCast:       false,
	PendingChoiceCopyTarget:    false,
	PendingChoiceCreatureType:  false,
	PendingChoiceColor:         false,
	// The attacker whose damage is being assigned was theirs, and it
	// left with them.
	PendingChoiceDamageAssignment: false,
	// Their triggers are dropped outright (CR 800.4d,
	// dropPendingTriggersForLocked), so there is nothing left to order
	// and nothing left to pick a mode for. mode_pick (#764, CR 603.3c)
	// looks like pick_target's twin — an object's choice that is not a
	// cost — and is `false` for the rule one number earlier: "if a
	// triggered ability that would be controlled by a player who has
	// left the game would be put onto the stack, it isn't put on the
	// stack". Choosing a mode for an ability that will never be
	// announced is choosing nothing.
	PendingChoiceTriggerOrder: false,
	PendingChoiceModePick:     false,
	// CR 310.9's protector is chosen by the battle's controller as it
	// enters; the battle left with them.
	PendingChoiceChooseProtector: false,
	// The flip belongs to its flipper, who is named by the frame the
	// reassignment deliberately does not touch.
	PendingChoiceCoinCall: false,
	// CR 726: the allowance is on THEIR loop's tally key.
	PendingChoiceLoopShortcut: false,
	// confirm carries no record of whose material it is about —
	// QueueConfirmForEffect sets FromPlayer to the chooser
	// unconditionally — so the engine cannot tell a self-question from
	// one asked across the table. Dropped, as before #902. See the
	// ADR 0060 amendment's Consequences for the one-field change that
	// would lift this.
	PendingChoiceConfirm: false,

	// --- the CR 616 pair, which settles its own drop (#808) ------
	PendingChoiceReplacementOrder:    false,
	PendingChoiceOptionalReplacement: false,
}

// ReassignableChoiceKinds lists every kind choiceReassignDecisions has
// a row for, so a test can hold that list against the kinds the gate
// classifies and fail on one that was never decided. Order is
// unspecified. Mirrors ClassifiedChoiceKinds (choice_gate.go).
func ReassignableChoiceKinds() []PendingChoiceKind {
	out := make([]PendingChoiceKind, 0, len(choiceReassignDecisions))
	for kind := range choiceReassignDecisions {
		out = append(out, kind)
	}
	return out
}

// reassignDepartedChoiceLocked is THE predicate behind "is this prompt
// reassigned or dropped", and the only place the question is answered.
// It reports whether `c` — a choice owed by a player who has just left
// — was handed to somebody else; false means the caller drops it the
// way it always did.
//
// Two gates, in this order:
//
//  1. The KIND. choiceReassignDecisions above.
//  2. The OBJECT. CR 800.4g reassigns a choice an OBJECT requires, so
//     there has to still be one: the prompt's Source must be findable
//     and controlled by a player still in the game. A prompt whose
//     object left with its controller — which after CR 800.4a is every
//     prompt the departed player's own card asked — has nothing left to
//     require it.
//     CR 800.4h's "a rule requires" arm reassigns nothing today, and
//     that is a statement about the engine rather than about the rule:
//     every rule-required prompt it has (the legend rule, a cleanup
//     discard, a mulligan) is a choice about the asked player's own
//     permanents or hand, so all of them are `false` above anyway.
//  3. The MATERIAL. A prompt about the departed player's OWN pool has
//     no answer that does anything: CR 800.4a took that pool out of the
//     game a moment ago, so every branch of Torment of Hailfire's "lose
//     3 life unless you sacrifice or discard" acts on a player who is
//     not there. FromPlayer is the engine's name for whose material a
//     prompt is about — protocol's redactChoiceCards reads the same
//     field for the same meaning — so FromPlayer == Chooser is the test.
//     An unset FromPlayer makes no claim about material and falls
//     through to the other two gates.
//
// Caller must hold g.mu.
func (g *Game) reassignDepartedChoiceLocked(c *PendingChoice) bool {
	if c == nil || !choiceReassignDecisions[c.Kind] {
		return false
	}
	if c.FromPlayer == c.Chooser {
		return false
	}
	controller := g.choiceObjectControllerLocked(c)
	if controller == uuid.Nil || controller == c.Chooser {
		return false
	}
	if p := g.playerByIDLocked(controller); p == nil || p.Eliminated {
		return false
	}
	to := g.choiceInheritorLocked(c, controller)
	if to == uuid.Nil {
		return false
	}
	return g.reassignChoiceLocked(c, to)
}

// choiceInheritorLocked is CR 800.4g's "who makes it instead", and the
// only inheritor policy in the tree. `controller` is the controller of
// the object that requires the choice, already checked to be a player
// still in the game. Returns uuid.Nil when nobody is left, which is the
// caller's signal to drop.
//
// ONE WALK, the seats in turn order starting after the seat that left:
//
//   - The FIRST surviving opponent of the controller wins. That is
//     CR 800.4g's second sentence — "if the original choice was to be
//     made by an opponent of the controller of the object, that player
//     chooses another opponent if possible" — and it is the sentence
//     that always fires, because a choice the controller themselves
//     owed cannot reach here: CR 800.4a would have taken the object out
//     of the game along with them.
//   - The controller is the "if possible" fallback, taken only when no
//     other opponent is left. That is the first sentence's plain
//     reading: the controller chooses another player, and the only
//     other player is themselves.
//
// The engine does not prompt the controller to NOMINATE somebody, which
// is what the rule literally describes. Turn order from the departed
// seat is the deterministic stand-in, for the reason every other
// "choose a player" default in this engine is APNAP: a table of four
// bots has to reach the same answer as a table of four humans, and an
// extra prompt to pick who answers a prompt is a second place the table
// can wedge.
//
// Caller must hold g.mu.
func (g *Game) choiceInheritorLocked(c *PendingChoice, controller uuid.UUID) uuid.UUID {
	if c == nil {
		return uuid.Nil
	}
	n := len(g.Seats)
	seat := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == c.Chooser {
			seat = i
			break
		}
	}
	if seat < 0 {
		return uuid.Nil
	}
	fallback := uuid.Nil
	for offset := 1; offset <= n; offset++ {
		p := g.Seats[(seat+offset)%n]
		if p == nil || p.Eliminated || p.ID == c.Chooser {
			continue
		}
		if p.ID == controller {
			if fallback == uuid.Nil {
				fallback = p.ID
			}
			continue
		}
		return p.ID
	}
	return fallback
}

// choiceObjectControllerLocked names the object that requires the
// choice — CR 800.4g's "the object" — or uuid.Nil when there isn't one
// the engine can still see, which is CR 800.4h's case.
//
// A prompt's Source is the card that asked. It is looked up across
// every zone rather than the battlefield alone, because a resolution-
// time prompt outlives its spell: the card is in its owner's graveyard
// by the time the question is answered (pending_choice.go's header).
// A card off the battlefield may carry no controller, so ownership is
// the fallback — for a spell the two are the same player in every case
// this reaches.
//
// Caller must hold g.mu.
func (g *Game) choiceObjectControllerLocked(c *PendingChoice) uuid.UUID {
	if c == nil || c.Source == uuid.Nil {
		return uuid.Nil
	}
	card := g.findCardByIDLocked(c.Source)
	if card == nil {
		return uuid.Nil
	}
	if card.Controller != uuid.Nil {
		return card.Controller
	}
	return card.Owner
}
