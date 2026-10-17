package game

import "github.com/google/uuid"

// initiative.go turns the initiative from a marker a player toggles by
// hand into the designation CR 726 describes (ADR 0203, #4096).
//
//	726.1. The initiative is a designation a player can have. There is
//	       no initiative in a game until an effect instructs a player
//	       to take the initiative.
//	726.2. There are three inherent triggered abilities associated
//	       with having the initiative. These triggered abilities have
//	       no source and are controlled by the player who had the
//	       initiative at the time the abilities triggered. This is an
//	       exception to rule 113.8. The full text of these abilities
//	       are "At the beginning of the upkeep of the player who has
//	       the initiative, that player ventures into Undercity,"
//	       "Whenever one or more creatures a player controls deal
//	       combat damage to the player who has the initiative, the
//	       controller of those creatures takes the initiative," and
//	       "Whenever a player takes the initiative, that player
//	       ventures into Undercity."
//	726.3. Only one player can have the initiative at a time. As a
//	       player takes the initiative, the player who currently has
//	       the initiative ceases to have it.
//	726.4. If the player who has the initiative leaves the game, the
//	       active player takes the initiative at the same time that
//	       player leaves the game. If the active player is leaving the
//	       game or if there is no active player, the next player in
//	       turn order takes the initiative.
//	726.5. If the player who currently has the initiative is
//	       instructed to take the initiative, this causes the last
//	       triggered ability in 726.2 to trigger but does not create a
//	       second initiative designation.
//
// (Quoted from the September 25, 2026 edition, the one AGENTS.md §6
// pins. The issue that asked for this cited CR 725; in this edition
// 725 is the monarch and the initiative is 726.)
//
// The shape is the monarch's (monarch.go): the three abilities have
// no source card for the harvester to find, so they ride the listener
// registry, and each queues a sourceless StackItem with a keyed body
// so the table gets its response window and a table with one waiting
// is a restore point.

// EventInitiativeTaken — a player took the initiative (CR 726). Actor
// is the player who took it, Target the player who had it before
// (uuid.Nil when nobody did; equal to Actor when the holder took it
// again, CR 726.5), Source the object whose effect made them take it
// (uuid.Nil for the combat-damage trigger and CR 726.4's hand-on).
//
// Emitted on EVERY take, not only on a change of holder, because CR
// 726.5 makes taking it again trigger the venture. Also emitted with
// Actor uuid.Nil when the sandbox clears the marker by hand, so the
// layer pass re-reads; nothing triggers on that.
const EventInitiativeTaken EventKind = "initiative_taken"

// initiativeTriggers is the Listener for CR 726.2's three abilities
// and CR 726.4's hand-on. Registered from NewGame beside
// monarchTriggers.
type initiativeTriggers struct{}

// OnEvent runs under g.mu in write mode from notifyListenersLocked.
func (initiativeTriggers) OnEvent(g *Game, ev Event) {
	switch ev.Kind {
	case EventInitiativeTaken:
		if ev.Actor != uuid.Nil {
			g.queueInitiativeVentureLocked(ev.Actor, "takes the initiative")
		}
	case EventBeginUpkeep:
		if g.Initiative != uuid.Nil && ev.Actor == g.Initiative {
			g.queueInitiativeVentureLocked(g.Initiative, "upkeep")
		}
	case EventDealDamage:
		g.initiativeCombatDamageTriggerLocked(ev)
	case EventPlayerEliminated:
		g.initiativeLeftTheGameLocked()
	}
}

// queueInitiativeVentureLocked queues "that player ventures into
// Undercity" (CR 726.2's first and third abilities), controlled by
// `player` — who has the initiative as each one triggers.
//
// Caller must hold g.mu in write mode.
func (g *Game) queueInitiativeVentureLocked(player uuid.UUID, why string) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return
	}
	label := "the initiative — " + p.Name + " ventures into Undercity"
	if why == "upkeep" {
		label = "the initiative — at the beginning of " + p.Name + "'s upkeep, they venture into Undercity"
	}
	params := EffectParams{Player: player}
	g.queueHarvestedTriggerLocked(&StackItem{
		Kind:       StackItemTriggered,
		Controller: player,
		Owner:      player,
		Label:      label,
		Body:       initiativeVentureBody.Key(),
		Params:     params,
		Effect:     bodyEffect(initiativeVentureBody.Key(), params),
	})
}

// initiativeVentureBody is "initiative/venture": Params.Player — the
// player the ability was about when it triggered, not whoever has the
// initiative now — ventures into Undercity (CR 701.49d).
var initiativeVentureBody = DelayedBody("initiative/venture", func(g *Game, _ *StackItem, p EffectParams) error {
	return g.VentureForEffect(p.Player, uuid.Nil, UndercityKey)
})

// initiativeCombatDamageTriggerLocked is CR 726.2's second ability:
// "Whenever one or more creatures a player controls deal combat damage
// to the player who has the initiative, the controller of those
// creatures takes the initiative."
//
// ONE trigger per batch, not one per creature: combat damage is dealt
// simultaneously (CR 510.2), and the engine emits one EventDealDamage
// per creature, so the second and later events of the same damage step
// find the first trigger still waiting on PendingTriggers and add
// nothing (the triggerAlreadyPendingFrom dedup the batched-damage cards
// use). A first-strike step and the regular step are two batches, with
// priority between them, so a double striker that connects twice takes
// it twice — and ventures twice.
//
// Caller must hold g.mu in write mode.
func (g *Game) initiativeCombatDamageTriggerLocked(ev Event) {
	holder := g.Initiative
	if !ev.Combat || holder == uuid.Nil || ev.Target != holder {
		return
	}
	// Event.Actor on a combat damage event is the source's controller,
	// snapshotted as the damage was dealt (monarch.go has the reasons).
	taker := ev.Actor
	if taker == uuid.Nil || taker == holder {
		return
	}
	if src := findBattlefieldCard(g, ev.Source); src != nil && !src.IsCreature() {
		return
	}
	claimant := g.playerByIDLocked(taker)
	if claimant == nil || claimant.Eliminated {
		return
	}
	for _, it := range g.PendingTriggers {
		if it != nil && it.Body == initiativeTakeBody.Key() && it.Params.Player == taker && it.Controller == holder {
			return
		}
	}
	params := EffectParams{Player: taker}
	g.queueHarvestedTriggerLocked(&StackItem{
		Kind: StackItemTriggered,
		// CR 726.2: controlled by the player who HAD the initiative
		// when it triggered — the one losing it.
		Controller: holder,
		Owner:      holder,
		Label:      "the initiative — " + claimant.Name + " takes the initiative",
		Body:       initiativeTakeBody.Key(),
		Params:     params,
		Effect:     bodyEffect(initiativeTakeBody.Key(), params),
	})
}

// initiativeTakeBody is "initiative/take": Params.Player takes the
// initiative, which triggers their venture.
var initiativeTakeBody = DelayedBody("initiative/take", func(g *Game, _ *StackItem, p EffectParams) error {
	g.takeInitiativeLocked(p.Player, uuid.Nil)
	return nil
})

// initiativeLeftTheGameLocked is CR 726.4: when the player who has the
// initiative leaves, the active player takes it at the same time — or,
// if the active player is the one leaving, the next player in turn
// order. Called on EventPlayerEliminated, emitted after the turn
// cursor has walked off an eliminated active seat (monarch.go), so one
// walk from the active seat covers both halves.
//
// A take, so the new holder ventures (CR 726.2's third ability).
//
// Caller must hold g.mu in write mode.
func (g *Game) initiativeLeftTheGameLocked() {
	for _, p := range g.Seats {
		// The player who left takes their dungeon with them (CR 800.4a).
		if p != nil && p.Eliminated {
			p.Dungeon = nil
		}
	}
	if g.Initiative == uuid.Nil {
		return
	}
	if p := g.playerByIDLocked(g.Initiative); p != nil && !p.Eliminated {
		return
	}
	start := g.Turn.ActiveSeat
	if start < 0 || start >= len(g.Seats) {
		start = 0
	}
	for offset := 0; offset < len(g.Seats); offset++ {
		if s := g.Seats[(start+offset)%len(g.Seats)]; s != nil && !s.Eliminated {
			g.takeInitiativeLocked(s.ID, uuid.Nil)
			return
		}
	}
	g.Initiative = uuid.Nil
}

// takeInitiativeLocked is the one write of Game.Initiative that gives
// it to a player. Every route goes through it — a card's "you take the
// initiative" (TakeInitiativeForEffect), the combat-damage trigger,
// CR 726.4's hand-on and the sandbox's manual set_initiative — so every
// one of them triggers the venture. A player who has left the game
// does not take it (CR 800.4a).
//
// Caller must hold g.mu in write mode.
func (g *Game) takeInitiativeLocked(player, source uuid.UUID) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return
	}
	previous := g.Initiative
	g.Initiative = player
	g.EmitEvent(Event{
		Kind:   EventInitiativeTaken,
		Actor:  player,
		Target: previous,
		Source: source,
	})
}

// TakeInitiativeForEffect is "<player> takes the initiative" from
// inside a resolving spell or ability (the effect_api.go convention:
// the caller holds g.mu in write mode). `source` is the object whose
// effect it is. Returns ErrPlayerNotFound for an ID that names no
// seat; a seat that has left takes nothing (CR 800.4a).
func (g *Game) TakeInitiativeForEffect(player, source uuid.UUID) error {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return ErrPlayerNotFound
	}
	g.takeInitiativeLocked(player, source)
	return nil
}

// HasInitiative reports whether `player` has the initiative — "if you
// have the initiative". Caller must hold g.mu.
func HasInitiative(g *Game, player uuid.UUID) bool {
	return g != nil && player != uuid.Nil && g.Initiative == player
}
