package game

import (
	"strings"

	"github.com/google/uuid"
)

// turn_tally.go — what has happened this turn, counted once as it
// happens (#586, Discussion #559 item 1).
//
// Before this, "you gained life this turn", "a creature died this
// turn", "this ability already resolved this turn" and their
// relatives were each answered by walking g.Events backwards to the
// last EventBeginUpkeep and counting: about thirty near-identical
// helpers in the catalog, each O(events this turn) per evaluation and
// each blind to anything that happened in the untap step, which is
// before the upkeep event. The engine already kept three real
// per-turn tallies (SpellsCastThisTurn, LandsPlayedThisTurn,
// LoyaltyActivatedThisTurn); this is the rest of them, in one place.
//
// One built-in listener (turnTallyListener) bumps the counters from
// the same events every other consumer reads, and onTurnAdvanceLocked
// resets the whole thing where the other three reset. Nothing at an
// emit site changes. For the filtered long tail ("you sacrificed a
// FOOD this turn") EventsThisTurn returns the bounded slice of this
// turn's events, so a card-side scan is short and starts at the real
// turn boundary rather than at the upkeep.

// PlayerTurnTally is what one player has done, or had done to them,
// this turn.
type PlayerTurnTally struct {
	// LifeGained is the sum of positive life changes.
	LifeGained int `json:"lifeGained,omitempty"`
	// LifeLost is the sum of negative life changes plus damage dealt
	// to the player — the CR 119.3 reading every "lost life this
	// turn" card in the catalog already used.
	LifeLost int `json:"lifeLost,omitempty"`
	// CardsDrawn counts draws (one event per card).
	CardsDrawn int `json:"cardsDrawn,omitempty"`
	// CreaturesDied counts creatures this player controlled that went
	// from the battlefield to a graveyard (CR 700.4).
	CreaturesDied int `json:"creaturesDied,omitempty"`
	// TokensCreated counts tokens created under this player's control.
	TokensCreated int `json:"tokensCreated,omitempty"`
	// PermanentsSacrificed counts permanents this player sacrificed.
	PermanentsSacrificed int `json:"permanentsSacrificed,omitempty"`
	// LandsEntered counts lands that entered under this player's
	// control — landfall's "this turn" tally.
	LandsEntered int `json:"landsEntered,omitempty"`
	// AttacksDeclared counts creatures this player declared as
	// attackers.
	AttacksDeclared int `json:"attacksDeclared,omitempty"`
	// CombatDamageToPlayers is the combat damage this player's
	// creatures dealt to players.
	CombatDamageToPlayers int `json:"combatDamageToPlayers,omitempty"`
}

// TurnTally is the per-turn record on Game. Reset on turn advance.
type TurnTally struct {
	// Players holds each player's tally; a player with nothing
	// counted has no entry. Read through Game.TurnTallyFor.
	Players map[uuid.UUID]PlayerTurnTally `json:"players,omitempty"`
	// CreaturesDied is the table-wide count, any controller.
	CreaturesDied int `json:"creaturesDied,omitempty"`
	// Resolved counts stack items that resolved this turn, keyed by
	// TallyKey(source, label) — the "once per turn" gate for an
	// ability that must not fire twice.
	Resolved map[string]int `json:"resolved,omitempty"`
	// Triggered counts triggered abilities announced onto
	// PendingTriggers this turn, keyed the same way.
	Triggered map[string]int `json:"triggered,omitempty"`
	// FirstEvent is the index into Game.Events at which this turn
	// began; EventsThisTurn slices from it.
	FirstEvent int `json:"firstEvent,omitempty"`
}

// TallyKey names one printed ability for the Resolved / Triggered
// maps: the source permanent and the stack label it announces with.
func TallyKey(source uuid.UUID, label string) string {
	return source.String() + "|" + label
}

// TurnTallyFor returns playerID's tally for the current turn (the
// zero value when nothing has been counted). Caller must hold g.mu.
func (g *Game) TurnTallyFor(playerID uuid.UUID) PlayerTurnTally {
	return g.TurnTally.Players[playerID]
}

// ResolvedThisTurn reports how many times the ability labelled
// `label` from `source` has resolved this turn. An empty label sums
// every ability of the source. Caller must hold g.mu.
func (g *Game) ResolvedThisTurn(source uuid.UUID, label string) int {
	return sumTally(g.TurnTally.Resolved, source, label)
}

// TriggeredThisTurn reports how many times the ability labelled
// `label` from `source` has been announced this turn. An empty label
// sums every ability of the source. Caller must hold g.mu.
func (g *Game) TriggeredThisTurn(source uuid.UUID, label string) int {
	return sumTally(g.TurnTally.Triggered, source, label)
}

func sumTally(m map[string]int, source uuid.UUID, label string) int {
	if label != "" {
		return m[TallyKey(source, label)]
	}
	prefix := source.String() + "|"
	n := 0
	for k, v := range m {
		if strings.HasPrefix(k, prefix) {
			n += v
		}
	}
	return n
}

// EventsThisTurn returns the events emitted since the turn began, in
// order. The slice aliases Game.Events: read it, do not keep it
// across a mutation. Caller must hold g.mu.
func (g *Game) EventsThisTurn() []Event {
	start := g.TurnTally.FirstEvent
	if start > len(g.Events) {
		start = len(g.Events)
	}
	if start < 0 {
		start = 0
	}
	return g.Events[start:]
}

// resetTurnTallyLocked starts a fresh tally at the current end of the
// event log. Caller must hold g.mu.
func (g *Game) resetTurnTallyLocked() {
	g.TurnTally = TurnTally{FirstEvent: len(g.Events)}
}

// cloneTurnTally deep-copies the maps; the counters are plain values.
func cloneTurnTally(t TurnTally) TurnTally {
	out := TurnTally{CreaturesDied: t.CreaturesDied, FirstEvent: t.FirstEvent}
	if len(t.Players) > 0 {
		out.Players = make(map[uuid.UUID]PlayerTurnTally, len(t.Players))
		for k, v := range t.Players {
			out.Players[k] = v
		}
	}
	out.Resolved = copyStringIntMap(t.Resolved)
	out.Triggered = copyStringIntMap(t.Triggered)
	return out
}

// turnTallyListener is the built-in Listener that keeps TurnTally
// current. Registered in NewGame ahead of the layer listener and the
// trigger harvester, so it reads the last-known battlefield
// characteristics of a dying creature before harvestLTB drops them,
// and so an AppliesTo that asks "has this already happened this turn"
// during the SAME event sees the count as it stood before that event
// — which is what the log scans it replaces saw.
type turnTallyListener struct{}

func (turnTallyListener) OnEvent(g *Game, ev Event) {
	switch ev.Kind {
	case EventChangeLife:
		if ev.Target == uuid.Nil {
			return
		}
		g.bumpPlayerTally(ev.Target, func(p *PlayerTurnTally) {
			if ev.Amount > 0 {
				p.LifeGained += ev.Amount
			} else {
				p.LifeLost -= ev.Amount
			}
		})
	case EventDealDamage:
		if ev.Amount <= 0 || g.playerByIDLocked(ev.Target) == nil {
			return
		}
		g.bumpPlayerTally(ev.Target, func(p *PlayerTurnTally) { p.LifeLost += ev.Amount })
		if ev.Combat && ev.Actor != uuid.Nil {
			g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.CombatDamageToPlayers += ev.Amount })
		}
	case EventDrawCard:
		g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.CardsDrawn++ })
	case EventTokenCreated:
		g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.TokensCreated++ })
	case EventSacrifice:
		g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.PermanentsSacrificed++ })
	case EventAttack:
		g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.AttacksDeclared++ })
	case EventETB:
		if ev.CardID == uuid.Nil {
			return
		}
		if c := g.findCardByIDLocked(ev.CardID); c != nil && c.IsLand() {
			g.bumpPlayerTally(c.Controller, func(p *PlayerTurnTally) { p.LandsEntered++ })
		}
	case EventLTB:
		if ev.NewZone != ZoneGraveyard || ev.CardID == uuid.Nil {
			return
		}
		c := g.findCardByIDLocked(ev.CardID)
		if c == nil {
			return
		}
		wasCreature := c.IsCreature()
		if lki, ok := g.lastKnownBattlefield[ev.CardID]; ok {
			wasCreature = hasTypeFold(lki.Types, "creature")
		}
		if !wasCreature {
			return
		}
		g.TurnTally.CreaturesDied++
		controller := c.Controller
		if controller == uuid.Nil {
			controller = c.Owner
		}
		g.bumpPlayerTally(controller, func(p *PlayerTurnTally) { p.CreaturesDied++ })
	case EventResolve:
		if ev.Source == uuid.Nil {
			return
		}
		if g.TurnTally.Resolved == nil {
			g.TurnTally.Resolved = map[string]int{}
		}
		g.TurnTally.Resolved[TallyKey(ev.Source, ev.Label)]++
	case EventTrigger:
		if ev.Source == uuid.Nil {
			return
		}
		if g.TurnTally.Triggered == nil {
			g.TurnTally.Triggered = map[string]int{}
		}
		g.TurnTally.Triggered[TallyKey(ev.Source, ev.Label)]++
	}
}

// bumpPlayerTally applies fn to playerID's tally, creating the entry
// on first touch. A nil player ID (an actor-less admin or SBA event)
// is not a player and is not counted.
func (g *Game) bumpPlayerTally(playerID uuid.UUID, fn func(*PlayerTurnTally)) {
	if playerID == uuid.Nil {
		return
	}
	if g.TurnTally.Players == nil {
		g.TurnTally.Players = map[uuid.UUID]PlayerTurnTally{}
	}
	p := g.TurnTally.Players[playerID]
	fn(&p)
	g.TurnTally.Players[playerID] = p
}

func hasTypeFold(types []string, want string) bool {
	for _, t := range types {
		if strings.EqualFold(t, want) {
			return true
		}
	}
	return false
}
