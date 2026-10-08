package game

import (
	"strconv"
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
// Two dimensions the counters cannot carry are recorded AS THE EVENT
// HAPPENS rather than counted: EnteredSubtypes / SacrificedSubtypes /
// CombatDamagedPlayers read the permanent while it is still there to
// read, because "a Faerie you control dealt combat damage to that
// player this turn" is asked after the Faerie has traded and, if it
// was a token, ceased to exist (CR 704.5d). Entered is the third: a
// per-OBJECT cell for "did this permanent enter this turn".
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
	// LifeLost is the sum of negative life changes plus the life that
	// damage dealt to the player cost them (CR 119.2, CR 120.3a):
	// damage from a source with infect, or to a player whose life
	// total can't change, is dealt but loses no life and is not
	// counted (#2105, Event.DamageLifeLoss).
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
	// CardsDiscarded counts cards this player discarded (one event
	// per card, CR 701.9a), whatever the discard's cause or eventual
	// destination — a madness card exiled instead of binned, or one
	// redirected to the top of the library, still counts (#1112,
	// Change of Fortune's "for each card you've discarded this
	// turn").
	CardsDiscarded int `json:"cardsDiscarded,omitempty"`
	// ArtifactsOrCreaturesSacrificed counts the artifacts and creatures
	// this player sacrificed (ADR 0100 §3) — Dargo, the Shipwrecker's
	// "{2} less to cast for each other artifact or creature you've
	// sacrificed this turn". Read off the permanent as it is sacrificed,
	// while it is still on the battlefield (EventSacrifice fires then),
	// so an artifact creature counts once.
	ArtifactsOrCreaturesSacrificed int `json:"artifactsOrCreaturesSacrificed,omitempty"`
	// PermanentsLeft counts the permanents this player controlled that
	// left the battlefield, by any route and to any zone (revolt,
	// CR 702.136; #2148): a sacrifice, a bounce, an exile, a death, a
	// token that ceased to exist, a permanent that left and came back.
	// "As it left" is the controller the exit stamped on the event
	// (CR 603.10a), not the card's field afterwards. Read through
	// Game.PermanentLeftThisTurn.
	PermanentsLeft int `json:"permanentsLeft,omitempty"`
	// LandsToGraveyard counts the lands this player controlled that
	// were put into a graveyard from the battlefield (#2664, ADR 0049's
	// 2026-10-08 amendment): destroyed, sacrificed, or put there by any
	// other route. A land that was bounced or exiled does not count. The
	// permanent is read as it last existed (CR 608.2h), so an animated
	// land that dies counts. The Lady of Otaria's end-step "if" reads it
	// through Game.LandToGraveyardThisTurn. Additive within schema v7.
	LandsToGraveyard int `json:"landsToGraveyard,omitempty"`
	// LifeLow / LifeLowSet are this player's lowest life total this
	// turn, and PoisonHigh their highest poison count: the CR 732 loop
	// breaker's progress marks (#2450, ADR 0055's 2026-10-07 amendment,
	// option B). A new low or a new high restarts the loop runs, because
	// it can happen only so many times before CR 704.5a or CR 122.1f
	// takes the player out. Written by noteLoopLifeLowLocked and
	// noteLoopPoisonHighLocked and read by nothing else.
	LifeLow    int  `json:"lifeLow,omitempty"`
	LifeLowSet bool `json:"lifeLowSet,omitempty"`
	PoisonHigh int  `json:"poisonHigh,omitempty"`
	// EnergyPaidOrLost is the energy this player has paid or lost this
	// turn (ADR 0129 §6): the sum of every negative energy delta that
	// landed on them, whether it was paid (payEnergyLocked, CR 107.14)
	// or removed by an effect (Final Act). Read off
	// EventPlayerCounterPlaced, so a sandbox stepper that takes energy
	// off the seat counts too, as a sandbox life change counts toward
	// LifeLost. Izzet Generatorium's "Activate only if you've paid or
	// lost four or more {E} this turn" and Blaster Hulk's "{1} less for
	// each {E} you've paid or lost this turn" read it through
	// Game.EnergyPaidOrLostThisTurn. Additive within schema v7.
	EnergyPaidOrLost int `json:"energyPaidOrLost,omitempty"`
	// SpeedTriggered is CR 702.179d's "this ability triggers only once
	// each turn": set as this player's inherent speed trigger is
	// queued (speedTriggerLocked). ADR 0138. Additive within schema v7.
	SpeedTriggered bool `json:"speedTriggered,omitempty"`
}

// TurnTally is the per-turn record on Game. Reset on turn advance.
type TurnTally struct {
	// Players holds each player's tally; a player with nothing
	// counted has no entry. Read through Game.TurnTallyFor.
	Players map[uuid.UUID]PlayerTurnTally `json:"players,omitempty"`
	// CreaturesDied is the table-wide count, any controller.
	CreaturesDied int `json:"creaturesDied,omitempty"`
	// Resolved counts stack items that resolved this turn, keyed by
	// ObjectTallyKey(source, epoch, label) — the "once per turn" gate
	// for an ability that must not fire twice. PER OBJECT (#936,
	// CR 400.7): a permanent that left the battlefield and came back
	// is a new object and starts its own count, because the epoch in
	// the key changed with it. Read through Game.ResolvedThisTurn,
	// which takes the epoch of whichever object the source names now.
	Resolved map[string]int `json:"resolved,omitempty"`
	// Triggered counts triggered abilities announced onto
	// PendingTriggers this turn, keyed the same way and per object
	// for the same reason.
	Triggered map[string]int `json:"triggered,omitempty"`
	// EnteredSubtypes counts the permanents that entered the
	// battlefield this turn, per the player they entered under and per
	// subtype they had as they entered, keyed by subtypeTallyKey. A
	// permanent with every creature type (changeling, CR 702.73a) is
	// counted once under enteredAllCreatureTypes instead of once per
	// creature type. Read through Game.EnteredWithSubtypeThisTurn.
	// Recorded at the entry because the question is about the object
	// as it entered: one that has since died, or since gained or lost
	// a type, answers as it was then (#743, Lilypad Village).
	EnteredSubtypes map[string]int `json:"enteredSubtypes,omitempty"`
	// Entered counts the battlefield entries each permanent made this
	// turn, keyed by the instance ID the entry produced. Read through
	// Game.EnteredThisTurn.
	//
	// The OBJECT dimension EnteredSubtypes does not have, and the
	// reason both exist: "each green creature that entered this turn"
	// (Oran-Rief) and "another Human entered this turn" (Éowyn, which
	// has to know whether the source is one of the entries
	// EnteredSubtypes counted) are questions about a named permanent,
	// not about a type.
	//
	// The boundary is the REAL turn boundary (#1009). resetTurnTallyLocked
	// runs from onTurnBeganLocked, before the untap step, so a permanent
	// that entered during the untap step — an untap-step choice or an
	// untap-step trigger putting one onto the battlefield (ADR 0059
	// Decision 7, ADR 0070) — is counted. The event-log walk this
	// replaces stopped at EventBeginUpkeep and answered "no" for it.
	//
	// Counted per ENTRY, not per object: a permanent blinked twice in
	// one turn has two. Nothing reads the count today; it is an int
	// because every other cell of the tally is, and because "how many
	// times" is the question the next card asks.
	Entered map[uuid.UUID]int `json:"entered,omitempty"`
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
	// CombatDamagedPlayers records which players were dealt combat
	// damage this turn and by WHAT, keyed by combatDamageTallyKey:
	// one cell per (the dealing creature's controller, an identity it
	// had as it dealt the damage, the damaged player). The identity is
	// the creature's NAME (Trygon Predator's "that player") or one of
	// its SUBTYPES (Alela's "one or more Faeries you control"), tagged
	// apart so a name and a subtype that spell the same cannot
	// collide. Read through PlayersDealtCombatDamageThisTurnByName and
	// PlayersDealtCombatDamageThisTurnBySubtype.
	//
	// Recorded at the damage for SacrificedSubtypes' reason (#596),
	// and combat is where that reason bites hardest: combat damage
	// kills the creature that dealt it at the very next state-based
	// check, and CR 704.5d then takes a TOKEN out of the graveyard
	// altogether. The walk this replaces looked the dealer up wherever
	// it had landed, so a Faerie token that traded in combat silently
	// dropped the player it hit out of the set — the set that a
	// trigger's target predicate reads, because a target predicate is
	// not handed the trigger's event.
	CombatDamagedPlayers map[string]int `json:"combatDamagedPlayers,omitempty"`
	// DamageDealers records WHICH creature objects dealt damage to
	// which player this turn (#2149): one entry per (object, victim,
	// combat or not), in the order the damage was dealt. Read through
	// CreaturesThatDealtCombatDamageToThisTurn and its any-damage
	// sibling.
	//
	// The OBJECT dimension CombatDamagedPlayers does not have, and the
	// reason both exist: "each opponent sacrifices a creature that
	// dealt combat damage to you this turn" (Witch-king of Angmar) and
	// "destroy target creature that dealt damage to you this turn"
	// (Spear of Heliod) are questions about named permanents, and the
	// name/subtype cells cannot say which Goblin it was.
	//
	// Each entry carries the creature's ObjectEpoch as it dealt the
	// damage, so a creature that has since been flickered is a NEW
	// object (CR 400.7) and reads as one that did not
	// (ObjectDealtCombatDamageToPlayerThisTurn compares the epoch to
	// the object's current one). Recorded at the damage, like its
	// siblings (#596): combat damage kills the dealer at the next
	// state-based check, and the list outlives it. Additive: omitted
	// when empty, so an older snapshot reads back as no records.
	DamageDealers []DamageDealtRecord `json:"damageDealers,omitempty"`
	// LoopRun is Resolved restarted at every player decision: the
	// CR 732 loop breaker's count of how many times one ability has
	// resolved with nobody casting, activating, answering a prompt
	// or declaring a creature in between. Reset to nil by
	// notePlayerDecisionLocked; read by loopSuspectedLocked and
	// nothing else. See loop_breaker.go (#628).
	//
	// Keyed by TallyKey — the CARD, not the object (#936). A blink
	// loop is a loop however many objects it mints on the way round,
	// and this is the count that must NOT restart when the permanent
	// leaves and re-enters.
	LoopRun map[string]int `json:"loopRun,omitempty"`
	// LoopAllowance is the CR 732 shortcut the loop's controller
	// agreed to, keyed like LoopRun (per CARD): how many more resolutions of
	// that ability the table runs before the breaker asks again.
	// Written by ResolveLoopShortcut, decremented by
	// noteResolutionForLoopLocked, cleared by
	// notePlayerDecisionLocked with everything else. Nil is the
	// ordinary state: nobody has taken a shortcut this turn.
	//
	// It is the ONLY counter #804 adds. The detector is still
	// loopSuspectedLocked over LoopRun; this says whether the answer
	// it gives is news. See loop_breaker.go (#804).
	LoopAllowance map[string]int `json:"loopAllowance,omitempty"`
	// LoopLow and LoopLowSet are the lowest count of work left on the
	// stack (StackMeta plus PendingTriggers) seen as a TRIGGERED
	// ability began to resolve, since the last player decision (#2450,
	// ADR 0055's 2026-10-07 amendment, option A). A resolution that
	// finds the count strictly below the low is a batch draining, not a
	// loop, and restarts the runs (noteLoopProgressLocked). Cleared by
	// notePlayerDecisionLocked with everything else.
	LoopLow    int  `json:"loopLow,omitempty"`
	LoopLowSet bool `json:"loopLowSet,omitempty"`
	// LoopActivated is the key of the activated ability the last
	// decision activated (#810's notePlayerActivationLocked), whose run
	// a draining batch does NOT restart: an activation loop is fed one
	// activation at a time, and the triggers each activation causes may
	// well drain between activations. Empty when the last decision was
	// not an activation.
	LoopActivated string `json:"loopActivated,omitempty"`
	// LoopProgressed marks a new lowest life total or a new highest
	// poison count since the last triggered resolution (#2450, option
	// B). noteLoopWorkLocked applies it as the next triggered ability
	// begins to resolve, and clears it.
	LoopProgressed bool `json:"loopProgressed,omitempty"`
	// Casts is every spell cast this turn, TABLE-WIDE, in cast order:
	// the instance ID of each spell as it went on the stack. Read
	// through Game.SpellsCastBeforeThisTurn.
	//
	// The one cell of this struct that is a LIST rather than a count,
	// because the question it answers is about an object's POSITION.
	// Storm (CR 702.40a) copies a spell "for each other spell that was
	// cast before it this turn", and no total can say which casts came
	// before a named one. See ADR 0086.
	//
	// Three properties, each of them a thing a count or a log scan
	// gets wrong (ADR 0086 Decision 1):
	//
	//   - Table-wide. CR 702.40a says "each other spell", not "each
	//     other spell you cast". Game.SpellsCastThisTurn is the
	//     per-player tally and stays exactly what it was.
	//   - Recorded AT THE CAST, like EnteredSubtypes and
	//     SacrificedSubtypes above and for the same reason (#596): a
	//     spell that was cast and then COUNTERED was still cast, and
	//     by the time storm asks, the card may be in a graveyard, in
	//     exile or (a token copy) nowhere at all. An ID recorded here
	//     is immune to every one of those.
	//   - Spells only. Playing a land is a special action (CR 116.2a)
	//     and emits no EventCast, so the land branch of CastSpell
	//     never reaches this list; a spell COPY is created and not
	//     cast (CR 707.10) and emits no EventCast either, so a storm
	//     chain counts only the real casts, as printed.
	Casts []uuid.UUID `json:"casts,omitempty"`
	// FirstEvent is the index into Game.Events at which this turn
	// began; EventsThisTurn slices from it.
	FirstEvent int `json:"firstEvent,omitempty"`
	// MonarchAtStart is who was the monarch as this turn began (#1722)
	// — Knights of the Black Rose's "if you were the monarch as the
	// turn began". Read through Game.MonarchAsTurnBegan. A stamp taken
	// in resetTurnTallyLocked rather than a walk back through the event
	// log, so it clones, snapshots and undoes with the rest of the
	// turn's record. uuid.Nil when nobody was.
	MonarchAtStart uuid.UUID `json:"monarchAtStart,omitempty"`
	// ModesChosen is "choose one that hasn't been chosen this turn"
	// (ADR 0097, #1749): the option indexes each modal ability has
	// chosen this turn, ascending, keyed by ObjectTallyKey(source,
	// epoch, label) — PER OBJECT for #936's reason, so a permanent
	// that left and came back chooses afresh, and never per
	// controller, so a change of control keeps it (the Demonic Pact
	// ruling). Written when the mode is CHOSEN, not at resolution,
	// and flushed with the rest of the tally when the turn begins.
	// Read through Game.modesChosenLocked. See mode_memory.go.
	ModesChosen map[string][]int `json:"modesChosen,omitempty"`
	// Attacks is every attack declared this turn, in declaration order
	// (ADR 0059 Decision 8): one record per attacking creature per
	// declaration, so a creature that attacks in two combats has two.
	// Read through TimesAttackedThisTurn and its siblings — "untap all
	// creatures that attacked this turn" (Relentless Assault), "attacks
	// for the first time each turn" (Aurelia, the Warleader).
	Attacks []AttackRecord `json:"attacks,omitempty"`
	// Exerts is every exert this turn, in order (ADR 0130 §1, CR
	// 701.43): one record per exert, so a creature exerted in two
	// combats has two. Written only by exertLocked; read through
	// ExertedThisTurn (Combat Celebrant's "if this creature hasn't
	// been exerted this turn"). Additive within v7: omitted when empty.
	Exerts []ExertRecord `json:"exerts,omitempty"`
	// StepsBegun and PhasesBegun count the steps and phase families
	// that have begun this turn; PhaseStarted is the PhaseID of the
	// phase the last count opened. They are what Turn.StepOrdinal and
	// Turn.PhaseOrdinal are read from (turn_plan.go).
	StepsBegun   map[Step]int      `json:"stepsBegun,omitempty"`
	PhasesBegun  map[PhaseKind]int `json:"phasesBegun,omitempty"`
	PhaseStarted int               `json:"phaseStarted,omitempty"`
}

// AttackRecord is one creature declared as an attacker (CR 508.1).
// Epoch is the attacker's Card.ObjectEpoch at the declaration, so a
// creature that left the battlefield and came back is a new object
// with no attacks of its own (CR 400.7). PhaseID is the combat phase
// it attacked in.
type AttackRecord struct {
	Attacker uuid.UUID `json:"attacker"`
	Epoch    int       `json:"epoch,omitempty"`
	Defender uuid.UUID `json:"defender,omitempty"`
	PhaseID  int       `json:"phaseId,omitempty"`
}

// TimesAttackedThisTurn is how many times the OBJECT `cardID` names
// now has been declared as an attacker this turn — "whenever ~ attacks
// for the first time each turn" is this reading 1 as the trigger
// fires (the tally listener runs ahead of the trigger harvester).
//
// Caller must hold g.mu.
func (g *Game) TimesAttackedThisTurn(cardID uuid.UUID) int {
	if cardID == uuid.Nil || len(g.TurnTally.Attacks) == 0 {
		return 0
	}
	epoch := g.objectEpochLocked(cardID)
	n := 0
	for _, a := range g.TurnTally.Attacks {
		if a.Attacker == cardID && a.Epoch == epoch {
			n++
		}
	}
	return n
}

// AttackedThisTurn reports whether the object `cardID` names attacked
// this turn: "creatures that attacked this turn".
//
// Caller must hold g.mu.
func (g *Game) AttackedThisTurn(cardID uuid.UUID) bool {
	return g.TimesAttackedThisTurn(cardID) > 0
}

// ObjectAttackedThisTurn is AttackedThisTurn for an object named by
// its ref rather than by where its card is now — an intervening "if ~
// attacked this turn" rechecked as the ability resolves (CR 603.4),
// which reads the source's last known information once it has left the
// battlefield (CR 608.2h). Taigam, Ojutai Master.
//
// Caller must hold g.mu.
func (g *Game) ObjectAttackedThisTurn(ref ObjectRef) bool {
	if ref.ID == uuid.Nil {
		return false
	}
	for _, a := range g.TurnTally.Attacks {
		if a.Attacker == ref.ID && a.Epoch == ref.Epoch {
			return true
		}
	}
	return false
}

// AttackedPlayersThisTurn is the players the object `cardID` names has
// attacked this turn, first attack first and each once. An attack on a
// planeswalker or a battle is not an attack on a player.
//
// Caller must hold g.mu.
func (g *Game) AttackedPlayersThisTurn(cardID uuid.UUID) []uuid.UUID {
	if cardID == uuid.Nil {
		return nil
	}
	epoch := g.objectEpochLocked(cardID)
	var out []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, a := range g.TurnTally.Attacks {
		if a.Attacker != cardID || a.Epoch != epoch || seen[a.Defender] {
			continue
		}
		if g.playerByIDLocked(a.Defender) == nil {
			continue
		}
		seen[a.Defender] = true
		out = append(out, a.Defender)
	}
	return out
}

// TallyKey names one printed ability of one CARD: the source's
// instance ID and the stack label it announces with.
//
// This is the CARD-LEVEL projection of the key, and since #936 it has
// exactly two readers left — TurnTally.LoopRun and
// TurnTally.LoopAllowance, the CR 732 loop breaker's pair
// (loop_breaker.go) — plus oncePerBatchFired, which compares its
// entries against the live event batch and so cannot read a stale
// one (event_batch.go).
//
// The breaker wants this projection and not the per-object one: a
// blink loop leaves and re-enters the battlefield on every iteration,
// and a loop is a loop whichever object is running it. Keyed per
// object, its run would reset every iteration and the threshold would
// never be reached — the escape hatch would be created by exactly the
// loops the breaker exists for.
func TallyKey(source uuid.UUID, label string) string {
	return source.String() + "|" + label
}

// ObjectTallyKey names one printed ability of one OBJECT: the
// card-level key plus the epoch that changes when the card changes
// zones (CR 400.7, Card.ObjectEpoch).
//
// This is the projection the CARD-FACING "only once each turn" gates
// read — TurnTally.Resolved and TurnTally.Triggered, through
// ResolvedThisTurn / TriggeredThisTurn — because CR 400.7 makes a
// permanent that leaves and returns a new object, and "whenever …, if
// this is the first time this has happened this turn" is a question
// about the object. A Conclave Mentor blinked in response to its own
// trigger may trigger again; the entries the old object wrote are
// still in the map and are simply unreachable, which is what "no
// memory of its previous existence" means for a map that is flushed
// when the turn ends.
//
// One key with two projections, not two keys: they are the same
// (source, label) pair, and the epoch is the one dimension that says
// which of the two questions is being asked. Which reader uses which
// is the doc comment on each map.
func ObjectTallyKey(source uuid.UUID, epoch int, label string) string {
	return objectTallyPrefix(source, epoch) + label
}

// objectTallyPrefix is ObjectTallyKey with no label — what a
// label-less sum scans for. The epoch sits BEFORE the label
// separator so that prefix is per (object), not per (card).
func objectTallyPrefix(source uuid.UUID, epoch int) string {
	return source.String() + "#" + strconv.Itoa(epoch) + "|"
}

// objectTallyKeyLocked is ObjectTallyKey for the object `source`
// names right now: the epoch is read off the card wherever it
// currently is. A source the engine can no longer find (a token that
// has ceased to exist, CR 111.7) reads as epoch zero, which is a key
// nothing else writes.
//
// Caller must hold g.mu.
func (g *Game) objectTallyKeyLocked(source uuid.UUID, label string) string {
	return ObjectTallyKey(source, g.objectEpochLocked(source), label)
}

// objectEpochLocked returns the current object epoch of the card
// `source` names, or zero when it is nowhere findable. Caller must
// hold g.mu.
func (g *Game) objectEpochLocked(source uuid.UUID) int {
	if c := g.findCardByIDLocked(source); c != nil {
		return c.ObjectEpoch
	}
	return 0
}

// EnergyPaidOrLostThisTurn is how much energy playerID has paid or
// lost this turn (ADR 0129 §6, PlayerTurnTally.EnergyPaidOrLost).
// Caller must hold g.mu.
func (g *Game) EnergyPaidOrLostThisTurn(playerID uuid.UUID) int {
	return g.TurnTally.Players[playerID].EnergyPaidOrLost
}

// TurnTallyFor returns playerID's tally for the current turn (the
// zero value when nothing has been counted). Caller must hold g.mu.
func (g *Game) TurnTallyFor(playerID uuid.UUID) PlayerTurnTally {
	return g.TurnTally.Players[playerID]
}

// ResolvedThisTurn reports how many times the ability labelled
// `label` from `source` has resolved this turn. An empty label sums
// every ability of the source.
//
// Per OBJECT (#936, CR 400.7): the count belongs to the permanent
// that resolved it, so a source that has left the battlefield and
// come back this turn answers zero. Caller must hold g.mu.
func (g *Game) ResolvedThisTurn(source uuid.UUID, label string) int {
	return sumTally(g.TurnTally.Resolved, g.objectTallyPrefixLocked(source), label)
}

// TriggeredThisTurn reports how many times the ability labelled
// `label` from `source` has been announced this turn. An empty label
// sums every ability of the source. Per OBJECT, like its sibling
// above. Caller must hold g.mu.
func (g *Game) TriggeredThisTurn(source uuid.UUID, label string) int {
	return sumTally(g.TurnTally.Triggered, g.objectTallyPrefixLocked(source), label)
}

// objectTallyPrefixLocked is the per-object key prefix for whichever
// object `source` names right now. Caller must hold g.mu.
func (g *Game) objectTallyPrefixLocked(source uuid.UUID) string {
	return objectTallyPrefix(source, g.objectEpochLocked(source))
}

// sumTally reads one cell of a keyed tally, or sums every cell under
// `prefix` when the caller named no label. The prefix carries the
// whole identity of the thing being asked about — a card for
// TallyKey, an object for ObjectTallyKey — so this function does not
// need to know which projection it is walking.
func sumTally(m map[string]int, prefix, label string) int {
	if label != "" {
		return m[prefix+label]
	}
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

// EnteredThisTurn reports whether the permanent `cardID` names
// entered the battlefield this turn — Oran-Rief's "each green
// creature that entered this turn".
//
// A tally of ENTRY EVENTS and not a walk of the log (#1009): the
// walk it replaces stopped at EventBeginUpkeep, so a permanent that
// entered during the untap step, before that event, answered "no".
// This answers from the turn boundary the tally is reset at, which
// is the real one.
//
// It is also not summoning sickness: a creature that entered during
// an opponent's turn still carries Card.SummonedThisTurn on yours
// (CR 302.6), and it did not enter this turn.
//
// Caller must hold g.mu.
func (g *Game) EnteredThisTurn(cardID uuid.UUID) bool {
	return cardID != uuid.Nil && g.TurnTally.Entered[cardID] > 0
}

// PlayersDealtCombatDamageThisTurnByName is the set of players a
// creature NAMED `name` under `controller`'s control dealt combat
// damage to this turn — Trygon Predator's "target artifact or
// enchantment THAT PLAYER controls", read by a target predicate that
// is not handed the trigger's event.
//
// Caller must hold g.mu.
func (g *Game) PlayersDealtCombatDamageThisTurnByName(controller uuid.UUID, name string) map[uuid.UUID]bool {
	if name == "" {
		return map[uuid.UUID]bool{}
	}
	return g.combatDamagedPlayersLocked(controller, combatDamageNameKey(name), false)
}

// PlayersDealtCombatDamageThisTurnBySubtype is the same set for a
// SUBTYPE — Alela, Cunning Conqueror's "whenever one or more Faeries
// you control deal combat damage to a player, goad target creature
// that player controls". A creature that had every creature type as
// it dealt the damage (changeling, CR 702.73a) counts for any
// creature type.
//
// Caller must hold g.mu.
func (g *Game) PlayersDealtCombatDamageThisTurnBySubtype(controller uuid.UUID, subtype string) map[uuid.UUID]bool {
	if subtype == "" {
		return map[uuid.UUID]bool{}
	}
	return g.combatDamagedPlayersLocked(controller, combatDamageSubtypeKey(subtype), IsCreatureType(subtype))
}

// combatDamagedPlayersLocked collects the seated players with a
// non-empty cell under one identity, folding in the changeling
// bucket when the caller asked about a creature type. Seats are
// walked rather than the map parsed: the map's key is three fields
// glued together and the set of players is short and already to hand.
func (g *Game) combatDamagedPlayersLocked(controller uuid.UUID, identity string, foldChangeling bool) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	if controller == uuid.Nil || len(g.TurnTally.CombatDamagedPlayers) == 0 {
		return out
	}
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		n := g.TurnTally.CombatDamagedPlayers[combatDamageTallyKey(controller, identity, p.ID)]
		if foldChangeling {
			n += g.TurnTally.CombatDamagedPlayers[combatDamageTallyKey(controller, combatDamageSubtypeKey(enteredAllCreatureTypes), p.ID)]
		}
		if n > 0 {
			out[p.ID] = true
		}
	}
	return out
}

// DamageDealtRecord is one creature object having dealt damage to one
// player (#2149). Epoch is the dealer's Card.ObjectEpoch when it dealt
// the damage (CR 400.7). Combat marks combat damage (CR 510); a
// creature that dealt both kinds this turn has two records.
type DamageDealtRecord struct {
	Source uuid.UUID `json:"source"`
	Epoch  int       `json:"epoch,omitempty"`
	Victim uuid.UUID `json:"victim"`
	Combat bool      `json:"combat,omitempty"`
}

// recordDamageDealerLocked notes that the creature `dealer` dealt
// damage to `victim`, once per (object, victim, kind). Only a
// creature is recorded: "a creature that dealt damage to you" is not
// answered by a burn spell or a planeswalker's ability. Caller must
// hold g.mu.
func (g *Game) recordDamageDealerLocked(dealer *Card, victim uuid.UUID, combat bool) {
	if dealer == nil || victim == uuid.Nil || !dealer.IsCreature() {
		return
	}
	rec := DamageDealtRecord{Source: dealer.InstanceID, Epoch: dealer.ObjectEpoch, Victim: victim, Combat: combat}
	for _, r := range g.TurnTally.DamageDealers {
		if r == rec {
			return
		}
	}
	g.TurnTally.DamageDealers = append(g.TurnTally.DamageDealers, rec)
}

// creaturesThatDealtDamageToLocked is the objects recorded against
// `victim`, first to deal damage first and each once; combatOnly
// narrows to combat damage. Caller must hold g.mu.
func (g *Game) creaturesThatDealtDamageToLocked(victim uuid.UUID, combatOnly bool) []ObjectRef {
	if victim == uuid.Nil {
		return nil
	}
	var out []ObjectRef
	for _, r := range g.TurnTally.DamageDealers {
		if r.Victim != victim || (combatOnly && !r.Combat) {
			continue
		}
		ref := ObjectRef{ID: r.Source, Epoch: r.Epoch}
		dup := false
		for _, o := range out {
			if o == ref {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, ref)
		}
	}
	return out
}

// CreaturesThatDealtCombatDamageToThisTurn is the creature objects
// that dealt combat damage to `victim` this turn, as they were when
// they dealt it (#2149) — Witch-king of Angmar's "a creature that
// dealt combat damage to you this turn". The refs outlive the
// creatures: one that died, or a token that ceased to exist, is still
// listed, so a reader that wants only the ones still in play checks
// the object is there (ObjectDealtCombatDamageToPlayerThisTurn).
//
// Caller must hold g.mu.
func (g *Game) CreaturesThatDealtCombatDamageToThisTurn(victim uuid.UUID) []ObjectRef {
	return g.creaturesThatDealtDamageToLocked(victim, true)
}

// CreaturesThatDealtDamageToThisTurn is the same list for damage of
// any kind (combat or not): "destroy target creature that dealt
// damage to you this turn" (Spear of Heliod, Retaliate).
//
// Caller must hold g.mu.
func (g *Game) CreaturesThatDealtDamageToThisTurn(victim uuid.UUID) []ObjectRef {
	return g.creaturesThatDealtDamageToLocked(victim, false)
}

// ObjectDealtCombatDamageToPlayerThisTurn reports whether the object
// the card `cardID` names RIGHT NOW is one that dealt combat damage to
// `victim` this turn. A creature that left and came back is a new
// object (CR 400.7) and answers false, because its epoch is not the
// recorded one. A card that is nowhere answers false.
//
// Caller must hold g.mu.
func (g *Game) ObjectDealtCombatDamageToPlayerThisTurn(cardID, victim uuid.UUID) bool {
	return g.objectDealtDamageToPlayerLocked(cardID, victim, true)
}

// ObjectDealtDamageToPlayerThisTurn is the any-damage sibling of
// ObjectDealtCombatDamageToPlayerThisTurn.
//
// Caller must hold g.mu.
func (g *Game) ObjectDealtDamageToPlayerThisTurn(cardID, victim uuid.UUID) bool {
	return g.objectDealtDamageToPlayerLocked(cardID, victim, false)
}

func (g *Game) objectDealtDamageToPlayerLocked(cardID, victim uuid.UUID, combatOnly bool) bool {
	if cardID == uuid.Nil || victim == uuid.Nil || len(g.TurnTally.DamageDealers) == 0 {
		return false
	}
	c := g.findCardByIDLocked(cardID)
	if c == nil {
		return false
	}
	for _, r := range g.TurnTally.DamageDealers {
		if r.Source == cardID && r.Epoch == c.ObjectEpoch && r.Victim == victim && (!combatOnly || r.Combat) {
			return true
		}
	}
	return false
}

// combatDamageTallyKey names one (dealer's controller, dealer
// identity, damaged player) cell of TurnTally.CombatDamagedPlayers.
func combatDamageTallyKey(controller uuid.UUID, identity string, victim uuid.UUID) string {
	return controller.String() + "|" + identity + "|" + victim.String()
}

// combatDamageNameKey and combatDamageSubtypeKey are the two tagged
// identities a dealing creature is recorded under. The tag is what
// keeps a card named "Faerie Mastermind" and the subtype "Faerie" in
// different cells. Folded to lower case, as every name and subtype
// comparison in the engine is.
func combatDamageNameKey(name string) string { return "n:" + strings.ToLower(name) }

func combatDamageSubtypeKey(subtype string) string { return "s:" + strings.ToLower(subtype) }

// recordCombatDamageToPlayerLocked records one creature's combat
// damage to one player under every identity a reader may ask for: its
// name, and each subtype it has right now. A creature with every
// creature type is recorded once under enteredAllCreatureTypes
// instead of once per creature type, exactly as recordSubtypeTally
// does for entries.
func (g *Game) recordCombatDamageToPlayerLocked(controller uuid.UUID, dealer *Card, victim uuid.UUID) {
	if controller == uuid.Nil || dealer == nil || victim == uuid.Nil {
		return
	}
	bump := func(identity string) {
		if g.TurnTally.CombatDamagedPlayers == nil {
			g.TurnTally.CombatDamagedPlayers = map[string]int{}
		}
		g.TurnTally.CombatDamagedPlayers[combatDamageTallyKey(controller, identity, victim)]++
	}
	if dealer.Name != "" {
		bump(combatDamageNameKey(dealer.Name))
	}
	all := HasAllCreatureTypes(dealer)
	for _, s := range dealer.Effective().Subtypes {
		if all && IsCreatureType(s) {
			continue
		}
		bump(combatDamageSubtypeKey(s))
	}
	if all {
		bump(combatDamageSubtypeKey(enteredAllCreatureTypes))
	}
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

// SpellsCastBeforeThisTurn reports how many spells were cast this
// turn, by any player, BEFORE the spell `spellID` was cast — storm's
// count (CR 702.40a, "copy it for each other spell that was cast
// before it this turn"). See ADR 0086.
//
// It reads TurnTally.Casts, which is append-only within a turn, so a
// spell's index is frozen by its own cast and the answer does not
// move under a caller. A spell cast in RESPONSE to the storm trigger
// is appended AFTER the storm spell and leaves that index alone,
// which is the rule and the reason storm can read this at resolution
// (CR 608.2h) instead of capturing a number at announce.
//
// "Other" is free: the spell's own entry is at the index this
// returns, so it is never counted.
//
// The MOST RECENT cast of that ID, which matters for exactly one
// shape: a spell countered or bounced back to its owner's hand
// (Remand) and cast again the same turn is two casts of one card, and
// a card keeps its instance ID across the round trip — so the list
// holds the ID twice. The question is always about the later one. The
// earlier cast's own trigger cannot still be waiting for an answer:
// CR 603.3 puts it on the stack ABOVE its spell, so it resolves
// before anything can bounce that spell. Taking the first index
// instead would undercount every recast, which is the case that
// actually happens at a table.
//
// A spell this turn's casts do not name answers zero. That is the
// honest answer for the two ways it happens — a spell cast on an
// earlier turn, and a COPY, which was created and not cast
// (CR 707.10) — and it is what a storm trigger on a copy would want:
// a copy has no storm trigger, because nothing was cast.
//
// Caller must hold g.mu.
func (g *Game) SpellsCastBeforeThisTurn(spellID uuid.UUID) int {
	for i := len(g.TurnTally.Casts) - 1; i >= 0; i-- {
		if g.TurnTally.Casts[i] == spellID {
			return i
		}
	}
	return 0
}

// resetTurnTallyLocked starts a fresh tally at the current end of the
// event log. Caller must hold g.mu.
func (g *Game) resetTurnTallyLocked() {
	g.TurnTally = TurnTally{FirstEvent: len(g.Events), MonarchAtStart: g.Monarch}
	// #628: the loop notice is a claim about THIS turn's resolutions,
	// so it dies with the counts it was derived from. A table that
	// stepped its way past a loop by hand starts the next turn with
	// automatic passing live again.
	g.LoopNotice = nil
}

// cloneTurnTally deep-copies the maps; the counters are plain values.
func cloneTurnTally(t TurnTally) TurnTally {
	out := TurnTally{CreaturesDied: t.CreaturesDied, FirstEvent: t.FirstEvent, MonarchAtStart: t.MonarchAtStart}
	if len(t.Players) > 0 {
		out.Players = make(map[uuid.UUID]PlayerTurnTally, len(t.Players))
		for k, v := range t.Players {
			out.Players[k] = v
		}
	}
	out.Resolved = copyStringIntMap(t.Resolved)
	out.EnteredSubtypes = copyStringIntMap(t.EnteredSubtypes)
	out.Entered = copyUUIDIntMap(t.Entered)
	out.SacrificedSubtypes = copyStringIntMap(t.SacrificedSubtypes)
	out.CombatDamagedPlayers = copyStringIntMap(t.CombatDamagedPlayers)
	if len(t.DamageDealers) > 0 {
		out.DamageDealers = append([]DamageDealtRecord(nil), t.DamageDealers...)
	}
	out.Triggered = copyStringIntMap(t.Triggered)
	out.LoopRun = copyStringIntMap(t.LoopRun)
	out.LoopAllowance = copyStringIntMap(t.LoopAllowance)
	out.LoopLow, out.LoopLowSet, out.LoopActivated = t.LoopLow, t.LoopLowSet, t.LoopActivated
	out.LoopProgressed = t.LoopProgressed
	out.ModesChosen = copyModesChosen(t.ModesChosen)
	if len(t.Attacks) > 0 {
		out.Attacks = append([]AttackRecord(nil), t.Attacks...)
	}
	if len(t.Exerts) > 0 {
		out.Exerts = append([]ExertRecord(nil), t.Exerts...)
	}
	if len(t.StepsBegun) > 0 {
		out.StepsBegun = make(map[Step]int, len(t.StepsBegun))
		for k, v := range t.StepsBegun {
			out.StepsBegun[k] = v
		}
	}
	if len(t.PhasesBegun) > 0 {
		out.PhasesBegun = make(map[PhaseKind]int, len(t.PhasesBegun))
		for k, v := range t.PhasesBegun {
			out.PhasesBegun[k] = v
		}
	}
	out.PhaseStarted = t.PhaseStarted
	// #1238: a fresh backing array, not the same slice header. The
	// clone is an undo restore point and the live game keeps
	// appending to its own list; sharing the array would let a cast
	// made after the snapshot reach the snapshot's view of the turn.
	if len(t.Casts) > 0 {
		out.Casts = append([]uuid.UUID(nil), t.Casts...)
	}
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
		// #2105: the life the damage COST, not the damage — infect
		// damage and damage to a locked life total lose no life.
		if lost := ev.DamageLifeLoss(); lost > 0 {
			g.bumpPlayerTally(ev.Target, func(p *PlayerTurnTally) { p.LifeLost += lost })
		}
		// #2149: which creature OBJECT it was, for "a creature that
		// dealt (combat) damage to you this turn".
		g.recordDamageDealerLocked(g.findCardByIDLocked(ev.Source), ev.Target, ev.Combat)
		if ev.Combat && ev.Actor != uuid.Nil {
			g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.CombatDamageToPlayers += ev.Amount })
			// #596, and the reason EventSacrifice records below: WHAT
			// dealt the damage is readable here and nowhere later.
			// Combat damage kills the creature that dealt it at the
			// next state-based check, and a token is gone from every
			// zone one check after that — so "which players did my
			// Faeries hit this turn", asked at the trigger's target
			// prompt, has to be answered from a record taken now.
			if src := g.findCardByIDLocked(ev.Source); src != nil {
				g.recordCombatDamageToPlayerLocked(ev.Actor, src, ev.Target)
			}
		}
	case EventPlayerCounterPlaced:
		// ADR 0129 §6: energy paid or lost. The event carries the
		// delta that landed, so a removal from fewer counters than it
		// named counts only what came off.
		if ev.Label == CounterEnergy && ev.Amount < 0 && g.playerByIDLocked(ev.Target) != nil {
			g.bumpPlayerTally(ev.Target, func(p *PlayerTurnTally) { p.EnergyPaidOrLost -= ev.Amount })
		}
	case EventDrawCard:
		g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.CardsDrawn++ })
	case EventDiscardCard:
		g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.CardsDiscarded++ })
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
				if c.IsArtifact() || c.IsCreature() {
					g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.ArtifactsOrCreaturesSacrificed++ })
				}
			}
		}
	case EventAttack:
		g.bumpPlayerTally(ev.Actor, func(p *PlayerTurnTally) { p.AttacksDeclared++ })
		// ADR 0059 Decision 8: the per-object attack history. A fresh
		// slice every time, never an append into shared capacity: the
		// undo clone holds this backing array too (#1238's reason).
		if ev.CardID != uuid.Nil {
			rec := AttackRecord{
				Attacker: ev.CardID,
				Epoch:    g.objectEpochLocked(ev.CardID),
				Defender: ev.Target,
				PhaseID:  g.Turn.PhaseID,
			}
			attacks := make([]AttackRecord, len(g.TurnTally.Attacks), len(g.TurnTally.Attacks)+1)
			copy(attacks, g.TurnTally.Attacks)
			g.TurnTally.Attacks = append(attacks, rec)
		}
		g.notePlayerDecisionLocked()
	case EventCast, EventBlock:
		// #1238 / CR 702.40a: the turn's cast ORDER, table-wide. Taken
		// here rather than in CastSpell so it is one record off one
		// event, beside the rest of the tally, and so an engine-driven
		// cast (a cascade hit, a "you may cast it" grant) is counted
		// exactly like a hand cast — both emit this event and both are
		// spells that were cast. The listener is registered ahead of
		// the trigger harvester (see NewGame), so a spell's own entry
		// is already here when its storm trigger is built.
		//
		// The CardID guard is not defensive: EventBlock shares this
		// arm, and a hand-written EventCast in a test may carry only
		// an Actor.
		if ev.Kind == EventCast && ev.CardID != uuid.Nil {
			g.TurnTally.Casts = append(g.TurnTally.Casts, ev.CardID)
		}
		// #628: the OTHER reason these two share an arm, and until
		// #1238 the only one — they are decision events for the CR 732
		// loop breaker. Casting a spell and declaring a blocker are
		// both "a player did something other than pass", which
		// restarts the loop run.
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
		// #1009: the per-object "entered this turn" cell. Bumped off
		// the EVENT and before the lookup, so an entry whose object
		// the engine can no longer find is still an entry.
		if g.TurnTally.Entered == nil {
			g.TurnTally.Entered = map[uuid.UUID]int{}
		}
		g.TurnTally.Entered[ev.CardID]++
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
		if ev.CardID == uuid.Nil {
			return
		}
		// #2148: revolt. Every battlefield exit, whatever the
		// destination, recorded under the controller it left from.
		g.recordPermanentLeftLocked(ev)
		if ev.NewZone != ZoneGraveyard {
			return
		}
		c := g.findCardByIDLocked(ev.CardID)
		if c == nil {
			return
		}
		wasCreature := c.IsCreature()
		wasLand := c.IsLand()
		if lki, ok := g.lastKnownBattlefield[ev.CardID]; ok {
			wasCreature = hasTypeFold(lki.Types, "creature")
			wasLand = hasTypeFold(lki.Types, "land")
		}
		if wasLand {
			// #2664: a land put into a graveyard from the battlefield,
			// under the controller it had as it left (CR 603.10a).
			controller, known := ev.LeftUnderControlOf()
			if !known {
				controller = c.Controller
			}
			if controller == uuid.Nil {
				controller = c.Owner
			}
			g.bumpPlayerTally(controller, func(p *PlayerTurnTally) { p.LandsToGraveyard++ })
		}
		if !wasCreature {
			return
		}
		g.TurnTally.CreaturesDied++
		// #1682 / CR 603.10a: "died under your control" is the
		// controller the permanent had as it left, which the exit
		// stamped on the event — not the graveyard card's field, which
		// CR 108.4 says names no one.
		controller, known := ev.LeftUnderControlOf()
		if !known {
			controller = c.Controller
		}
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
		// #936: the card-facing gate is per OBJECT (CR 400.7) and the
		// loop breaker is per CARD. Two projections of one key, both
		// taken here, so the two questions cannot drift apart.
		g.TurnTally.Resolved[g.objectTallyKeyLocked(ev.Source, ev.Label)]++
		// #628: the same count, restarted at each player decision, is
		// the CR 732 loop breaker's input. See loop_breaker.go.
		g.noteResolutionForLoopLocked(ev, TallyKey(ev.Source, ev.Label))
	case EventTrigger:
		if ev.Source == uuid.Nil {
			return
		}
		if g.TurnTally.Triggered == nil {
			g.TurnTally.Triggered = map[string]int{}
		}
		g.TurnTally.Triggered[g.objectTallyKeyLocked(ev.Source, ev.Label)]++
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

// copyUUIDIntMap is copyStringIntMap for a map keyed by instance ID
// (TurnTally.Entered). Same contract: an empty map copies to nil, so
// a cloned tally with nothing counted is the zero value.
func copyUUIDIntMap(in map[uuid.UUID]int) map[uuid.UUID]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func hasTypeFold(types []string, want string) bool {
	for _, t := range types {
		if strings.EqualFold(t, want) {
			return true
		}
	}
	return false
}

// recordPermanentLeftLocked bumps PermanentsLeft for the controller the
// leaving permanent had as it left (CR 603.10a), falling back to the
// card's controller and then its owner for an unstamped event.
func (g *Game) recordPermanentLeftLocked(ev Event) {
	controller, known := ev.LeftUnderControlOf()
	if !known {
		if c := g.findCardByIDLocked(ev.CardID); c != nil {
			controller = c.Controller
			if controller == uuid.Nil {
				controller = c.Owner
			}
		}
	}
	if controller == uuid.Nil {
		controller = ev.Actor
	}
	g.bumpPlayerTally(controller, func(p *PlayerTurnTally) { p.PermanentsLeft++ })
}

// PermanentLeftThisTurn reports revolt (CR 702.136): whether a
// permanent `playerID` controlled left the battlefield this turn.
//
// Caller must hold g.mu.
func (g *Game) PermanentLeftThisTurn(playerID uuid.UUID) bool {
	return g.TurnTallyFor(playerID).PermanentsLeft > 0
}

// LandToGraveyardThisTurn reports whether a land `playerID` controlled
// was put into a graveyard from the battlefield this turn (#2664, The
// Lady of Otaria's intervening "if", CR 603.4).
//
// Caller must hold g.mu.
func (g *Game) LandToGraveyardThisTurn(playerID uuid.UUID) bool {
	return g.TurnTallyFor(playerID).LandsToGraveyard > 0
}
