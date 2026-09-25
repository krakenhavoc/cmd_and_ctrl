package game

import "github.com/google/uuid"

// monarch.go turns the monarch from a marker a player toggles by hand
// into the two triggered abilities the rules actually give it.
//
// CR 724 (the monarch), reproduced because both halves of this file
// are a transcription of it:
//
//	724.2. There are two inherent triggered abilities associated with
//	being the monarch. These triggered abilities have no source and
//	are controlled by the player who was the monarch at the time the
//	abilities triggered. The full texts of these abilities are "At the
//	beginning of the monarch's end step, that player draws a card" and
//	"Whenever a creature deals combat damage to the monarch, that
//	creature's controller becomes the monarch."
//
//	725.3. Only one player can be the monarch at a time. As a player
//	becomes the monarch, the current monarch ceases to be the monarch.
//
//	725.4. If the monarch leaves the game, the active player becomes
//	the monarch at the same time as that player leaves the game. If
//	the active player is leaving the game or if there is no active
//	player, the next player in turn order becomes the monarch. If no
//	player still in the game can become the monarch, the game
//	continues with no monarch.
//
// #375: before this file, SetMonarch assigned Game.Monarch and that
// was the whole mechanic. A player who had the monarchy kept it after
// being hit, and never drew for it — the two things the designation
// is FOR. The crown was a sticker.
//
// WHY A LISTENER AND NOT A CATALOG TRIGGER. The S19 harvester
// (triggers.go) walks the battlefield and asks each CARD for its
// declared TriggeredAbility entries. These two abilities have no
// source card to declare them — they belong to the game, not to a
// permanent — so there is nothing for the harvester to find. The
// registry the engine already has for "something that watches the
// event log and owes no card" is Listener (listeners.go), which is
// where turnTallyListener and layerVersionBump live. monarchTriggers
// is the third, registered from NewGame.
//
// The abilities still use the STACK: each one queues a StackItem onto
// PendingTriggers through queueHarvestedTriggerLocked, exactly as a
// harvested card trigger does, so the table gets its response window
// (the crown can be Stifled, and a player can act knowing the draw is
// about to happen). The items carry SourceCardID uuid.Nil because CR
// 724.2 says they have no source; StackOverlay already falls back to
// the label for a sourceless item, which is how delayed triggers with
// no source render today.
type monarchTriggers struct{}

// OnEvent is the Listener half. Runs under g.mu in write mode from
// notifyListenersLocked, so everything it calls is a *Locked helper.
func (monarchTriggers) OnEvent(g *Game, ev Event) {
	switch ev.Kind {
	case EventDealDamage:
		g.monarchCombatDamageTriggerLocked(ev)
	case EventBeginEndStep:
		g.monarchEndStepTriggerLocked(ev)
	case EventPlayerEliminated:
		g.monarchLeftTheGameLocked()
	}
}

// monarchCombatDamageTriggerLocked is CR 724.2's second ability:
// "whenever a creature deals combat damage to the monarch, that
// creature's controller becomes the monarch".
//
// It reads the event rather than the combat state on purpose. Every
// path that lands combat damage on a player — unblocked attackers,
// trample overflow, the CR 510.1c multi-blocker assignment prompt,
// and the CR 616 resume that finishes a damage event which paused
// mid-step — funnels through emitDealDamageLocked, and that is the
// one place Event.Combat is set. Watching the event is therefore the
// only way to see all of them at once, and it is why a first-strike
// hit and a regular hit both count.
//
// One trigger per creature that connects, which is what CR 725.3
// means by "as a player becomes the monarch, the current monarch
// ceases to be": two creatures under different controllers both
// connecting put two triggers on the stack, the monarch orders them
// (CR 603.3b), and the crown ends up with the controller of the one
// that resolves LAST. Two creatures under the SAME controller
// produce two items with the same label and no source, which
// seatNeedsTriggerOrder correctly treats as interchangeable — no
// ordering prompt for a choice with one outcome.
//
// Caller must hold g.mu in write mode.
func (g *Game) monarchCombatDamageTriggerLocked(ev Event) {
	if !ev.Combat || g.Monarch == uuid.Nil || ev.Target != g.Monarch {
		return
	}
	// Event.Actor on a combat damage event is the damage source's
	// CONTROLLER, snapshotted into the damage tail when the event was
	// created (see combatDamageTailLocked). That snapshot is what
	// "that creature's controller" has to mean: the attacker may have
	// died to blocker damage in the same substep, and a mind-control
	// effect may have ended, before the trigger resolves.
	newMonarch := ev.Actor
	if newMonarch == uuid.Nil || newMonarch == g.Monarch {
		return
	}
	// "Whenever a CREATURE deals combat damage." A source that is
	// still on the battlefield and is not a creature (a battle or a
	// planeswalker cannot deal combat damage, but the sandbox's
	// MarkCombatDamage verb can be pointed at anything) does not pass
	// the crown. A source that has already left is taken at its word:
	// only a creature was in combat to begin with.
	if src := findBattlefieldCard(g, ev.Source); src != nil && !src.IsCreature() {
		return
	}
	claimant := g.playerByIDLocked(newMonarch)
	if claimant == nil || claimant.Eliminated {
		return
	}
	params := EffectParams{Player: newMonarch}
	g.queueHarvestedTriggerLocked(&StackItem{
		Kind: StackItemTriggered,
		// CR 724.2: controlled by the player who WAS the monarch when
		// the ability triggered — the one losing the crown, not the
		// one taking it. That is what makes CR 800.4a correct when
		// the hit is lethal: the trigger leaves with its controller
		// and the crown is handed on by CR 725.4 below instead.
		Controller: g.Monarch,
		Owner:      g.Monarch,
		Label:      "the monarch — " + claimant.Name + " becomes the monarch",
		// ADR 0041 P9 (#1497, tier 4): the monarch's two triggers have
		// no catalog row and no source card, so they are keyed
		// directly. The new monarch — captured before as `newMonarch`
		// — rides Params.Player instead.
		Body:   monarchCrownBody.Key(),
		Params: params,
		Effect: bodyEffect(monarchCrownBody.Key(), params),
	})
}

// monarchCrownBody is "monarch/crown" (ADR 0041 P9, #1497, tier 4):
// Params.Player becomes the monarch.
var monarchCrownBody = DelayedBody("monarch/crown", func(g *Game, _ *StackItem, p EffectParams) error {
	g.becomeMonarchLocked(p.Player)
	return nil
})

// monarchEndStepTriggerLocked is CR 724.2's first ability: "at the
// beginning of the monarch's end step, that player draws a card".
//
// THE MONARCH'S end step, not every end step: the draw happens once a
// turn cycle at most, on the turn of whoever holds the crown when
// that turn's end step begins. EventBeginEndStep's Actor is the
// active player, so the gate is one comparison — and it is the same
// gate every "at the beginning of your end step" catalog card uses.
//
// Caller must hold g.mu in write mode.
func (g *Game) monarchEndStepTriggerLocked(ev Event) {
	if g.Monarch == uuid.Nil || ev.Actor != g.Monarch {
		return
	}
	monarch := g.playerByIDLocked(g.Monarch)
	if monarch == nil || monarch.Eliminated {
		return
	}
	crowned := g.Monarch
	params := EffectParams{Player: crowned}
	g.queueHarvestedTriggerLocked(&StackItem{
		Kind:       StackItemTriggered,
		Controller: crowned,
		Owner:      crowned,
		Label:      "the monarch — draw a card",
		// ADR 0041 P9 (#1497, tier 4): keyed directly, with "that
		// player" — the monarch when the ability triggered, not
		// whoever holds the crown at resolution — carried as
		// Params.Player rather than captured.
		Body:   monarchDrawBody.Key(),
		Params: params,
		Effect: bodyEffect(monarchDrawBody.Key(), params),
	})
}

// monarchDrawBody is "monarch/draw" (ADR 0041 P9, #1497, tier 4):
// Params.Player draws a card. Losing the crown in response to your own
// end-step trigger does not cost you the card, because Params.Player
// is fixed when the trigger was created, not re-read off g.Monarch.
var monarchDrawBody = DelayedBody("monarch/draw", func(g *Game, _ *StackItem, p EffectParams) error {
	return g.DrawNForEffect(p.Player, 1)
})

// monarchLeftTheGameLocked is CR 725.4: the crown never falls off the
// table. When the monarch leaves, the active player takes it; if the
// active player is the one leaving, the next player in turn order
// does.
//
// Called on EventPlayerEliminated, which eliminatePlayerLocked emits
// AFTER advancePastEliminatedLocked has already walked the cursor off
// an eliminated active seat. So "the active player" here is by
// construction a player still in the game, and the two halves of
// CR 725.4 collapse into one lookup.
//
// This is a reassignment by game rule, not a triggered ability: it
// happens immediately, not on the stack ("at the same time as that
// player leaves the game").
//
// Caller must hold g.mu in write mode.
func (g *Game) monarchLeftTheGameLocked() {
	if g.Monarch == uuid.Nil {
		return
	}
	if p := g.playerByIDLocked(g.Monarch); p != nil && !p.Eliminated {
		// Somebody else left. The monarch is unaffected.
		return
	}
	start := g.Turn.ActiveSeat
	if start < 0 || start >= len(g.Seats) {
		start = 0
	}
	for offset := 0; offset < len(g.Seats); offset++ {
		if s := g.Seats[(start+offset)%len(g.Seats)]; s != nil && !s.Eliminated {
			g.becomeMonarchLocked(s.ID)
			return
		}
	}
	// "If no player still in the game can become the monarch, the
	// game continues with no monarch."
	g.becomeMonarchLocked(uuid.Nil)
}

// becomeMonarchLocked is the one write to Game.Monarch that the
// rules half of the mechanic goes through. Refuses to crown a player
// who is no longer seated or who has been eliminated between the
// trigger and its resolution; uuid.Nil clears the designation.
//
// Caller must hold g.mu in write mode.
func (g *Game) becomeMonarchLocked(playerID uuid.UUID) {
	if playerID != uuid.Nil {
		p := g.playerByIDLocked(playerID)
		if p == nil || p.Eliminated {
			return
		}
	}
	g.Monarch = playerID
}
