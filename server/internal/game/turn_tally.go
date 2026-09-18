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
// the same events every other consumer reads, and onTurnBeganLocked
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
	// EnteredSubtypes counts the permanents that entered the
	// battlefield this turn, per the player they entered under and per
	// subtype they had as they entered, keyed by enteredSubtypeKey. A
	// permanent with every creature type (changeling, CR 702.73a) is
	// counted once under enteredAllCreatureTypes instead of once per
	// creature type. Read through Game.EnteredWithSubtypeThisTurn.
	// Recorded at the entry because the question is about the object
	// as it entered: one that has since died, or since gained or lost
	// a type, answers as it was then (#743, Lilypad Village).
	EnteredSubtypes map[string]int `json:"enteredSubtypes,omitempty"`
	// SacrificedSubtypes is the same tally for the other end of a
	// permanent's life: the permanents each player SACRIFICED this
	// turn, per subtype they had as they were sacrificed. Read
	// through Game.SacrificedWithSubtypeThisTurn.
	//
	// Recorded at the sacrifice for the same reason (#596): the
	// object is read while it is still on the battlefield, because by
	// the time anything asks, it may not be anywhere. "You sacrificed
	// a Food this turn" used to be answered by walking the event log
	// and looking the sacrificed card up wherever it had landed, and
	// a Food is a TOKEN — CR 704.5d removes it from the graveyard at
	// the next state-based check, so the lookup found nothing and the
	// condition silently stopped being true for the commonest case
	// the card was printed for.
	SacrificedSubtypes map[string]int `json:"sacrificedSubtypes,omitempty"`
	// LoopRun is Resolved restarted at every player decision: the
	// CR 726 loop breaker's count of how many times one ability has
	// resolved with nobody casting, activating, answering a prompt
	// or declaring a creature in between. Reset to nil by
	// notePlayerDecisionLocked; read by loopSuspectedLocked and
	// nothing else. See loop_breaker.go (#628).
	LoopRun map[string]int `json:"loopRun,omitempty"`
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

// enteredAllCreatureTypes is the EnteredSubtypes key suffix for a
// permanent that entered with every creature type. Not a subtype
// anyone can print, so it never collides with a real one.
const enteredAllCreatureTypes = "*"

// subtypeTallyKey names one (player, subtype) cell of
// TurnTally.EnteredSubtypes or TurnTally.SacrificedSubtypes. Subtypes
// are compared case-insensitively, as everywhere else in the engine.
func subtypeTallyKey(player uuid.UUID, subtype string) string {
	return player.String() + "|" + strings.ToLower(subtype)
}

// EnteredWithSubtypeThisTurn reports how many permanents entered the
// battlefield under playerID's control this turn having `subtype` as
// they entered — "if a Bird, Frog, Otter, or Rat entered the
// battlefield under your control this turn" (Lilypad Village), "for
// each Goblin that entered under your control this turn". A permanent
// that had every creature type counts for any creature type.
//
// "As they entered" is the characteristics the permanent carried at
// its EventETB: its printed values, with a copy effect it entered
// under (CR 707.2) and the face it entered on already applied. The
// layer cache is rebuilt after an entry, not during it, so continuous
// effects from static abilities are NOT part of that reading. That
// makes a type GRANTED at the moment of entry invisible — a Soldier
// that entered while its controller's Maskwood Nexus was on the
// battlefield counts as a Soldier only, where the printed rules count
// it as every type — which is weaker than the rules, and a card that
// reads this declares it. The opposite case, a static that REMOVES a
// subtype from a permanent as it enters, would read stronger; no
// catalog card removes a creature type from anything but the
// permanent an Aura is already attached to, or from itself only while
// a condition holds (Arixmethes, The Warring Triad), and a reader for
// those types has to account for it.
//
// Caller must hold g.mu.
func (g *Game) EnteredWithSubtypeThisTurn(playerID uuid.UUID, subtype string) int {
	if playerID == uuid.Nil || subtype == "" {
		return 0
	}
	return subtypeTallyCount(g.TurnTally.EnteredSubtypes, playerID, subtype)
}

// SacrificedWithSubtypeThisTurn reports how many permanents playerID
// sacrificed this turn having `subtype` as they were sacrificed — "if
// you sacrificed a Food this turn" (Elanor Gardner), "for each
// Treasure you sacrificed this turn". A permanent that had every
// creature type counts for any creature type.
//
// "As they were sacrificed" is the reading its EnteredSubtypes twin
// documents, and the reason this is a tally rather than an event-log
// walk is #596: a sacrificed TOKEN — which is what most of the cards
// asking this question sacrifice — does not survive in the graveyard
// for a later lookup to read.
//
// Caller must hold g.mu.
func (g *Game) SacrificedWithSubtypeThisTurn(playerID uuid.UUID, subtype string) int {
	if playerID == uuid.Nil || subtype == "" {
		return 0
	}
	return subtypeTallyCount(g.TurnTally.SacrificedSubtypes, playerID, subtype)
}

// subtypeTallyCount reads one (player, subtype) cell, folding in the
// changeling bucket for a creature type.
func subtypeTallyCount(m map[string]int, playerID uuid.UUID, subtype string) int {
	n := m[subtypeTallyKey(playerID, subtype)]
	if IsCreatureType(subtype) {
		n += m[subtypeTallyKey(playerID, enteredAllCreatureTypes)]
	}
	return n
}

// recordEnteredSubtypesLocked adds one entering permanent to
// TurnTally.EnteredSubtypes under `controller`.
func (g *Game) recordEnteredSubtypesLocked(controller uuid.UUID, c *Card) {
	recordSubtypeTally(&g.TurnTally.EnteredSubtypes, controller, c)
}

// recordSacrificedSubtypesLocked adds one sacrificed permanent to
// TurnTally.SacrificedSubtypes under `actor`, the player who
// sacrificed it.
func (g *Game) recordSacrificedSubtypesLocked(actor uuid.UUID, c *Card) {
	recordSubtypeTally(&g.TurnTally.SacrificedSubtypes, actor, c)
}

// recordSubtypeTally bumps one cell per subtype the permanent has
// right now, allocating the map on first use. A permanent with every
// creature type (changeling, CR 702.73a) is counted once under
// enteredAllCreatureTypes instead of once per creature type.
func recordSubtypeTally(m *map[string]int, playerID uuid.UUID, c *Card) {
	if playerID == uuid.Nil || c == nil {
		return
	}
	all := HasAllCreatureTypes(c)
	bump := func(subtype string) {
		if *m == nil {
			*m = map[string]int{}
		}
		(*m)[subtypeTallyKey(playerID, subtype)]++
	}
	for _, s := range c.Effective().Subtypes {
		if all && IsCreatureType(s) {
			continue
		}
		bump(s)
	}
	if all {
		bump(enteredAllCreatureTypes)
	}
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
	// #628: the loop notice is a claim about THIS turn's resolutions,
	// so it dies with the counts it was derived from. A table that
	// stepped its way past a loop by hand starts the next turn with
	// automatic passing live again.
	g.LoopNotice = nil
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
	out.EnteredSubtypes = copyStringIntMap(t.EnteredSubtypes)
	out.SacrificedSubtypes = copyStringIntMap(t.SacrificedSubtypes)
	out.Triggered = copyStringIntMap(t.Triggered)
	out.LoopRun = copyStringIntMap(t.LoopRun)
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
		// #596: EventSacrifice fires while the permanent is still on
		// the battlefield (sacrifice.go), which is the only moment a
		// sacrificed TOKEN can be read — CR 704.5d takes it out of
		// the graveyard at the next state-based check. Record what it
		// was now; "you sacrificed a Food this turn" is asked later.
		if ev.CardID != uuid.Nil && ev.Actor != uuid.Nil {
			if c := g.findCardByIDLocked(ev.CardID); c != nil {
				g.recordSacrificedSubtypesLocked(ev.Actor, c)
			}
		}
	case EventAttack:
		g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.AttacksDeclared++ })
		g.notePlayerDecisionLocked()
	case EventCast, EventBlock:
		// #628: these two bump no counter — they are here only as
		// decision events for the CR 726 loop breaker. Casting a
		// spell and declaring a blocker are both "a player did
		// something other than pass", which restarts the loop run.
		// Activations and answered prompts have no event of their own
		// and notch notePlayerDecisionLocked directly; see
		// loop_breaker.go.
		//
		// An EventCast the ENGINE produced (a trigger that casts a
		// card) restarts the run too, which is over-clearing. That is
		// the safe direction: the cost is a loop that takes longer to
		// notice, never a breaker that fires on a real turn.
		g.notePlayerDecisionLocked()
	case EventETB:
		if ev.CardID == uuid.Nil {
			return
		}
		c := g.findCardByIDLocked(ev.CardID)
		if c == nil {
			return
		}
		if c.IsLand() {
			g.bumpPlayerTally(c.Controller, func(p *PlayerTurnTally) { p.LandsEntered++ })
		}
		controller := c.Controller
		if controller == uuid.Nil {
			controller = ev.Actor
		}
		g.recordEnteredSubtypesLocked(controller, c)
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
		key := TallyKey(ev.Source, ev.Label)
		if g.TurnTally.Resolved == nil {
			g.TurnTally.Resolved = map[string]int{}
		}
		g.TurnTally.Resolved[key]++
		// #628: the same count, restarted at each player decision, is
		// the CR 726 loop breaker's input. See loop_breaker.go.
		g.noteResolutionForLoopLocked(ev, key)
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
