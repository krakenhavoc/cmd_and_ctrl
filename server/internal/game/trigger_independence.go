package game

import (
	"reflect"
	"sort"

	"github.com/google/uuid"
)

// trigger_independence.go — "Ask only when the order matters" skips the
// CR 603.3b prompt for a batch whose triggers cannot change what each
// other does (#2884, ADR 0018's #2884 amendment).
//
// Two items are INDEPENDENT when, given the targets each was put on the
// stack with (CR 603.3d), nothing one of them writes is read or written
// by the other, and nothing on the board reacts to what either of them
// does. A batch whose items are pairwise independent goes on the stack
// in the order it was collected, the order every skipped batch uses.
//
// The check reads only declared data: each catalog row's Footprint,
// which the registry derives from an effects.Do of primitives that each
// declare a step (cards/effects/footprint.go). Anything it cannot read
// keeps the prompt. The direction is the one #1968 chose: a batch left
// asking costs a click; a batch wrongly skipped takes a real choice away.

// FootprintKind names one declared step of a triggered row's effect.
type FootprintKind string

const (
	// FootprintGainLife: the item's controller gains N life, or, with
	// AmountFromSubject, life equal to a characteristic of the trigger's
	// subject.
	FootprintGainLife FootprintKind = "gain_life"
	// FootprintDamageEachOpponent: the item's source deals N damage to
	// each opponent of the item's controller.
	FootprintDamageEachOpponent FootprintKind = "damage_each_opponent"
	// FootprintLoseLifeEachOpponent: each opponent of the item's
	// controller still in the game loses N life. Life loss, not damage,
	// and it names no damage source, so it has no lifelink, infect,
	// wither or toxic question (#3013).
	FootprintLoseLifeEachOpponent FootprintKind = "lose_life_each_opponent"
	// FootprintDamageController: the item's source deals N damage to the
	// item's controller (City of Brass, Midnight Reaper). The same damage
	// source questions as FootprintDamageEachOpponent.
	FootprintDamageController FootprintKind = "damage_controller"
	// FootprintTargetPlayerLosesLife: the player the item targets loses
	// N life (Blood Artist's and Falkenrath Noble's half of a drain).
	// Life loss, named to no damage source, like
	// FootprintLoseLifeEachOpponent.
	FootprintTargetPlayerLosesLife FootprintKind = "target_player_loses_life"
	// FootprintCounterOnSource: N counters of kind Counter on the item's
	// source, if it is still on the battlefield.
	FootprintCounterOnSource FootprintKind = "counter_on_source"
	// FootprintExileTargets: exile the card the item targets.
	FootprintExileTargets FootprintKind = "exile_targets"
	// FootprintDraw: the controller draws N cards.
	FootprintDraw FootprintKind = "draw"
	// FootprintMill: the controller mills N cards.
	FootprintMill FootprintKind = "mill"
	// FootprintScry and FootprintSurveil: the controller scries or
	// surveils N, with nothing hanging off the result.
	FootprintScry    FootprintKind = "scry"
	FootprintSurveil FootprintKind = "surveil"
	// FootprintEnergy: the controller gets N energy counters.
	FootprintEnergy FootprintKind = "energy"
	// FootprintCreateToken: the controller creates N copies of Token.
	FootprintCreateToken FootprintKind = "create_token"
	// FootprintMonarch: the controller becomes the monarch.
	FootprintMonarch FootprintKind = "monarch"
	// FootprintPumpSource: the item's source gets +N/+N until end of
	// turn, a layer-7c record pinned to that object, if it is still the
	// object that triggered and still on the battlefield. Declared for
	// an engine keyword trigger's body (keyedBodyFootprints), not by a
	// catalog primitive: prowess (#3013).
	FootprintPumpSource FootprintKind = "pump_source"
	// FootprintCondition: an intervening "if" re-checked as the item
	// resolves (CR 603.4). It changes nothing and emits nothing; it
	// declares the state its condition (Cond) reads, and the other steps
	// of the footprint are what the item does when the condition holds
	// (#3033). Every step a condition guards is declared as if it runs,
	// the conservative reading of "maybe".
	FootprintCondition FootprintKind = "condition"
)

// FootprintCond names the state an intervening "if" reads (#3033). The
// check reads each one as addConditionReadLocked says, and nothing else:
// a condition joins only with the argument made there for what can
// change it while the batch resolves.
type FootprintCond string

const (
	// FootprintCondLifeGained: the item's controller gained N or more
	// life this turn (The Gaffer). The turn's tally only rises.
	FootprintCondLifeGained FootprintCond = "life_gained"
	// FootprintCondLifeLost: the item's controller lost N or more life
	// this turn (Lunar Convocation's Bat). The tally only rises.
	FootprintCondLifeLost FootprintCond = "life_lost"
	// FootprintCondCreatureDied: a creature died under the item's
	// controller's control this turn (Faramir). The tally only rises.
	FootprintCondCreatureDied FootprintCond = "creature_died"
	// FootprintCondCitysBlessing: the item's controller has the city's
	// blessing, which is never lost once had (CR 702.131c).
	FootprintCondCitysBlessing FootprintCond = "citys_blessing"
	// FootprintCondNoSpellsCast: the item's controller cast no spells
	// this turn (Edgar). No step in the vocabulary casts a spell.
	FootprintCondNoSpellsCast FootprintCond = "no_spells_cast"
	// FootprintCondMonarch: the item's controller is the monarch (Ginger).
	FootprintCondMonarch FootprintCond = "monarch"
	// FootprintCondSource: a fact about the item's source object, such as
	// its counters (Runaway Steam-Kin's "fewer than three").
	FootprintCondSource FootprintCond = "source"
	// FootprintCondBoard: a fact about the permanents on the battlefield,
	// read through the layers (a creature with power 4 or greater, three
	// artifacts, seven land names, thirty counters).
	FootprintCondBoard FootprintCond = "board"
)

// keyedBodyFootprints is the footprint of each engine keyword trigger
// whose item is keyed by its stack body rather than stamped with a
// catalog row (ADR 0018's #3013 amendment, sub-PR 3). The check looks
// an item up here by its Body the way it looks a stamped item's row up
// in the catalog. A body joins only with the argument the amendment
// makes for prowess, and with a case in
// TestKeyedBodyFootprintsDoOnlyWhatTheyDeclare, which resolves the real
// body and fails on anything it changes or emits outside the
// declaration.
var keyedBodyFootprints = map[string][]FootprintStep{
	// CR 702.108a: "this creature gets +1/+1 until end of turn"
	// (resolveProwess). It reads only whether its source is still the
	// object that triggered, writes only a +1/+1 pinned to it, and
	// registering the record emits no event.
	prowessPumpBody.key: {{Kind: FootprintPumpSource, N: 1}},
}

// KeyedBodyFootprint is the declared footprint of an engine keyword
// trigger's stack body, for tests.
func KeyedBodyFootprint(body string) []FootprintStep {
	return keyedBodyFootprints[body]
}

// KeyedFootprintBodies lists the bodies keyedBodyFootprints declares,
// for the conformance test.
func KeyedFootprintBodies() []string {
	out := make([]string, 0, len(keyedBodyFootprints))
	for k := range keyedBodyFootprints {
		out = append(out, k)
	}
	return out
}

// FootprintStep is one declared step: a kind and its constants. Every
// object and player it touches is relative to the resolving item (its
// controller, its source, its chosen targets), never a constant ID.
type FootprintStep struct {
	Kind FootprintKind
	N    int
	// Counter is the counter kind of a FootprintCounterOnSource step.
	Counter string
	// TokenKey is the catalog key of a FootprintCreateToken step's
	// template ("" for one with no entry), and TokenManual whether the
	// template carries printed text the engine leaves to the table. A
	// key, not the template, so the step is plain data.
	TokenKey    string
	TokenManual bool
	// TokenKeywords are the template's own keywords. A keyword the
	// engine turns into a trigger (evolve, exalted, prowess…) gives the
	// token an ability no catalog row declares, so the check reads them
	// for what such a trigger watches (#3013).
	TokenKeywords []string
	// AmountFromSubject marks a FootprintGainLife step whose amount is
	// read at resolution off the trigger's subject, the object its
	// event names (CR 608.2h: "gain life equal to its toughness"). The
	// check counts that object as read while it is on the battlefield;
	// once it has left, the amount is its frozen last-known value and
	// nothing another item does can change it (#3013).
	AmountFromSubject bool
	// Cond is the state a FootprintCondition step reads, with N its
	// threshold where it has one.
	Cond FootprintCond
}

// FootprintTokenOf describes a token template for a FootprintStep.
func FootprintTokenOf(tmpl Card) (key string, manual bool, keywords []string) {
	return catalogKeyOf(&tmpl), Unimplemented(tmpl), append([]string(nil), tmpl.Keywords...)
}

// lifeDirection records which ways a batch item moves one player's life.
type lifeDirection uint8

const (
	lifeDown lifeDirection = 1 << iota
	lifeUp
)

// triggerAccess is one item's footprint resolved against its item and
// the board: what it reads and writes, and which events it causes.
type triggerAccess struct {
	controller uuid.UUID
	// source is the item's source card, for the sentence that says why
	// the prompt was asked (trigger_order_why.go). It never feeds a
	// decision.
	source uuid.UUID
	// Objects (cards and permanents), by instance ID.
	reads, writes map[uuid.UUID]bool
	// pumped are the objects the item pumps (FootprintPumpSource): a
	// layer-7c power and toughness record. markRulesLocked folds them
	// into writes; one the item changes in no other way sets pumpRules
	// rather than rules when it bears rules (#3013).
	pumped map[uuid.UUID]bool
	// Players.
	life     map[uuid.UUID]lifeDirection
	lifeLoss map[uuid.UUID]int
	cards    map[uuid.UUID]bool // library, hand and graveyard order
	// beyondDraw are the players in cards whose library the item does
	// more to than draw from the top: a mill, scry or surveil. Two items
	// that only draw from one library give the same hand and library
	// in either order; anything else that touches it does not.
	beyondDraw  map[uuid.UUID]bool
	draws       map[uuid.UUID]int
	playerReads map[uuid.UUID]bool // a player target
	// monarch is the one game-wide designation a step writes.
	monarch bool
	// rules is set when the item changes an object whose abilities
	// shape the rules every other item resolves under (a static
	// ability, a replacement effect, an attachment, a record that
	// names it). Everything reads the rules, so it conflicts with all.
	rules bool
	// pumpRules is set when the item pumps an object that bears rules.
	// A pump changes nothing about the object but its power and
	// toughness, so the only way it reaches past the object is a static
	// of the object's that reads them. That can change what another
	// item reads through the layers (charReads, traitReads), and
	// nothing else (#3013).
	pumpRules bool
	// traitReads is set when the item reads what the layers make of an
	// object other than its toughness: a damage source's lifelink,
	// infect, wither or toxic, or a chosen target's legality (CR
	// 608.2b reads its characteristics).
	traitReads bool
	// tokenWatches are the event kinds the keyword triggers of a token
	// the item creates watch. Such a token enters partway through the
	// batch, so it sees the events of the items after it and not of
	// those before (#3013).
	tokenWatches map[EventKind]bool
	// charReads is set when the item reads a characteristic of an
	// object on the battlefield through the layers (an amount read off
	// the trigger's subject, #3013): a static on the board can change
	// it when another item changes what that static reads.
	charReads bool
	// boardReads is set when the item's intervening "if" reads the
	// battlefield (FootprintCondBoard): another item that changes a
	// permanent, puts one onto or off the battlefield, or changes what a
	// static on the board reads can change whether the item does
	// anything (#3033).
	boardReads bool
	// monarchRead is set when the item's intervening "if" asks whether
	// its controller is the monarch (#3033).
	monarchRead bool
	emits       map[EventKind]bool
}

// writesAnything reports whether the item changes any game state a
// static ability's output could depend on.
func (a *triggerAccess) writesAnything() bool {
	return len(a.writes) > 0 || len(a.life) > 0 || len(a.cards) > 0 || a.monarch ||
		a.emits[EventTokenCreated] || a.emits[EventPlayerCounterPlaced] || a.emits[EventZoneMove]
}

// writesOnlyLife reports whether the only state the item changes is
// life totals (a gain, a loss, damage to players).
func (a *triggerAccess) writesOnlyLife() bool {
	return len(a.life) > 0 && len(a.writes) == 0 && len(a.cards) == 0 && !a.monarch &&
		!a.emits[EventTokenCreated] && !a.emits[EventPlayerCounterPlaced] && !a.emits[EventZoneMove]
}

// layerStatics is what the board's static abilities can do to a
// characteristic an item reads: whether any static can change power or
// toughness at all, and whether any of those reads a life total.
type layerStatics struct {
	pt, life bool
	// ptCard and lifeCard name the first card behind each flag, for
	// the sentence that says why the prompt was asked.
	ptCard, lifeCard string
	// anyLife and anyCards are set when any static, in any layer, reads
	// a life total, or a hand or the top of a library: what a board
	// condition can see change when another item only moves life or
	// cards (#3033).
	anyLife, anyCards bool
}

// layerStaticsLocked reads the battlefield and the emblems for static
// abilities that can change a creature's power or toughness: layer 7,
// a copy effect (layer 1), or an ability removal that could take away
// a characteristic-defining ability. The layer engine's contract
// (layerVersionBump) is that such a static's output depends on the
// battlefield, counters, taps and the turn, plus a life total or a
// hand only when it declares so; the check leans on the first half to
// stay conservative and on DependsOnLifeTotal to let a batch that only
// moves life totals through.
//
// Caller must hold g.mu.
func (g *Game) layerStaticsLocked() layerStatics {
	var out layerStatics
	zones := []*Zone{g.Battlefield}
	for _, p := range g.Seats {
		if p != nil {
			zones = append(zones, p.Emblems)
		}
	}
	for _, z := range zones {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			d := catalogDef(catalogAbilityKeyOf(&z.Cards[i]))
			if d == nil {
				continue
			}
			for _, st := range d.Static {
				out.anyLife = out.anyLife || st.DependsOnLifeTotal
				out.anyCards = out.anyCards || st.DependsOnHandSize || st.DependsOnLibraryTop
				if st.Layer == Layer7PT || st.Layer == Layer1Copy || st.RemovesAbilities {
					if !out.pt {
						out.ptCard = whyCardName(z.Cards[i])
					}
					out.pt = true
					if st.DependsOnLifeTotal {
						if !out.life {
							out.lifeCard = whyCardName(z.Cards[i])
						}
						out.life = true
					}
				}
			}
		}
	}
	return out
}

func newTriggerAccess(controller uuid.UUID) triggerAccess {
	return triggerAccess{
		controller:   controller,
		reads:        map[uuid.UUID]bool{},
		writes:       map[uuid.UUID]bool{},
		pumped:       map[uuid.UUID]bool{},
		tokenWatches: map[EventKind]bool{},
		life:         map[uuid.UUID]lifeDirection{},
		lifeLoss:     map[uuid.UUID]int{},
		cards:        map[uuid.UUID]bool{},
		beyondDraw:   map[uuid.UUID]bool{},
		draws:        map[uuid.UUID]int{},
		playerReads:  map[uuid.UUID]bool{},
		emits:        map[EventKind]bool{},
	}
}

func (a *triggerAccess) emit(kinds ...EventKind) {
	for _, k := range kinds {
		a.emits[k] = true
	}
}

// TriggerOrderReason says why a seat was asked to order its triggers
// (ADR 0018's #3013 amendment, sub-PR 1). It is the first thing the
// skips found that could make the order show, and it is what
// cmdctrl_trigger_order_prompts_total counts by. Every value is in
// TriggerOrderReasons, and metrics' closed label set for the counter
// holds each one (TestTriggerOrderReasonsAreMetricLabels).
type TriggerOrderReason string

const (
	// TriggerOrderReasonAlways: the seat chose Always ask (#1530, #1968).
	TriggerOrderReasonAlways TriggerOrderReason = "always"
	// TriggerOrderReasonUndescribed: an item the check cannot describe:
	// a row with no footprint (a hand-written closure, a "you may", a
	// mode clause, a clause built from the trigger), a Build whose item
	// is not plain, an item with Params, a payload, X or a division of
	// its own, or an item that is not a stamped catalog trigger.
	TriggerOrderReasonUndescribed TriggerOrderReason = "undescribed"
	// TriggerOrderReasonBoardOpaque: triggerOrderBoardOpaqueLocked gave
	// up on the board (a permanent whose text the engine leaves to the
	// table, a rules-bearing slot, a state trigger, a player static, a
	// continuing effect that is more than a characteristic change).
	TriggerOrderReasonBoardOpaque TriggerOrderReason = "board_opaque"
	// TriggerOrderReasonReactor: something in the game watches an event
	// kind the batch causes (somethingReactsLocked).
	TriggerOrderReasonReactor TriggerOrderReason = "reactor"
	// TriggerOrderReasonRipple: an object the batch writes is named by a
	// delayed trigger, a duration or another card's record
	// (writtenObjectsRippleLocked).
	TriggerOrderReasonRipple TriggerOrderReason = "ripple"
	// TriggerOrderReasonElimination: a player the batch could take out
	// of the game is its controller, or is touched by another item.
	TriggerOrderReasonElimination TriggerOrderReason = "elimination"
	// TriggerOrderReasonObjectTie: two items touch one object and at
	// least one of them writes it.
	TriggerOrderReasonObjectTie TriggerOrderReason = "object_tie"
	// TriggerOrderReasonRules: an item writes an object whose abilities
	// shape the rules the other items resolve under.
	TriggerOrderReasonRules TriggerOrderReason = "rules"
	// TriggerOrderReasonLibrary: two items both touch one player's
	// library or graveyard.
	TriggerOrderReasonLibrary TriggerOrderReason = "library"
	// TriggerOrderReasonMonarch: two items both make their controller
	// the monarch.
	TriggerOrderReasonMonarch TriggerOrderReason = "monarch"
	// TriggerOrderReasonCharacteristic: an item reads a characteristic
	// through the layers (an amount equal to the trigger subject's
	// toughness), and another item changes something a static on the
	// board could read to compute it (#3013).
	TriggerOrderReasonCharacteristic TriggerOrderReason = "characteristic"
	// TriggerOrderReasonCondition: an item's intervening "if" reads state
	// another item in the batch can change (#3033).
	TriggerOrderReasonCondition TriggerOrderReason = "condition"
)

// TriggerOrderReasons is every TriggerOrderReason, in a fixed order.
func TriggerOrderReasons() []TriggerOrderReason {
	return []TriggerOrderReason{
		TriggerOrderReasonAlways, TriggerOrderReasonUndescribed, TriggerOrderReasonBoardOpaque,
		TriggerOrderReasonReactor, TriggerOrderReasonRipple, TriggerOrderReasonElimination,
		TriggerOrderReasonObjectTie, TriggerOrderReasonRules, TriggerOrderReasonLibrary,
		TriggerOrderReasonMonarch, TriggerOrderReasonCharacteristic, TriggerOrderReasonCondition,
	}
}

// seatNeedsTriggerOrderLocked is seatNeedsTriggerOrder with the board
// in hand: in "when it matters" mode (and for a mode this binary does
// not know, which behaves as the default) a batch the pure skips would
// ask about is also skipped when its items are pairwise independent.
//
// Caller must hold g.mu.
func (g *Game) seatNeedsTriggerOrderLocked(items []*StackItem, mode TriggerOrderMode) bool {
	ask, _ := g.triggerOrderAskLocked(items, mode)
	return ask
}

// triggerOrderAskLocked is seatNeedsTriggerOrderLocked with the reason
// it asked. The reason is empty when it does not ask, and it never
// changes whether it asks.
//
// Caller must hold g.mu.
func (g *Game) triggerOrderAskLocked(items []*StackItem, mode TriggerOrderMode) (bool, TriggerOrderReason) {
	ask, why := g.triggerOrderAskWhyLocked(items, mode)
	return ask, why.reason
}

// triggerOrderAskWhyLocked is triggerOrderAskLocked with the facts the
// prompt's sentence names (trigger_order_why.go). Same decision, same
// reason: the extra detail is read off what the check already found.
//
// Caller must hold g.mu.
func (g *Game) triggerOrderAskWhyLocked(items []*StackItem, mode TriggerOrderMode) (bool, triggerOrderWhy) {
	if !seatNeedsTriggerOrder(items, mode) {
		// The one skip the board is asked about again: an identical
		// batch whose items read their triggering subjects (#3021).
		if w, shows := g.identicalBatchOrderShowsLocked(items, mode); shows {
			return true, w
		}
		return false, triggerOrderWhy{}
	}
	if mode == TriggerOrderAlways || mode == TriggerOrderNever {
		return true, triggerOrderWhy{reason: TriggerOrderReasonAlways}
	}
	if w := g.triggerOrderConflictLocked(items); w.reason != "" {
		return true, w
	}
	return false, triggerOrderWhy{}
}

// triggerOrderConflictLocked reports "" when every order of items gives
// the same game: each item is described by its row's footprint, nothing
// on the board reacts to what any of them does, and no two of them
// touch the same thing in a way whose order shows. Otherwise it returns
// the first reason it found that the order could show. (Before #3013
// this was triggersIndependentLocked, a bool.)
//
// Caller must hold g.mu.
func (g *Game) triggerOrderConflictLocked(items []*StackItem) triggerOrderWhy {
	if len(items) < 2 {
		return triggerOrderWhy{reason: TriggerOrderReasonUndescribed}
	}
	if card, opaque := g.triggerOrderBoardOpaqueLocked(); opaque {
		return triggerOrderWhy{reason: TriggerOrderReasonBoardOpaque, card: card}
	}
	accs := make([]triggerAccess, len(items))
	for i, t := range items {
		a, ok := g.triggerAccessLocked(t)
		if !ok {
			return triggerOrderWhy{reason: TriggerOrderReasonUndescribed, card: g.whyCardNameLocked(t.SourceCardID)}
		}
		a.source = t.SourceCardID
		accs[i] = a
	}
	emits := map[EventKind]bool{}
	written := map[uuid.UUID]bool{}
	loss := map[uuid.UUID]int{}
	draws := map[uuid.UUID]int{}
	for _, a := range accs {
		for k := range a.emits {
			emits[k] = true
		}
		for id := range a.writes {
			written[id] = true
		}
		for p, n := range a.lifeLoss {
			loss[p] += n
		}
		for p, n := range a.draws {
			draws[p] += n
		}
	}
	if card, kind, reacts := g.reactorLocked(emits); reacts {
		return triggerOrderWhy{reason: TriggerOrderReasonReactor, card: card, event: kind}
	}
	// A token the batch creates with a keyword trigger that watches an
	// event the batch causes (evolve watches entering) reacts to the
	// items after its maker and not to those before it.
	for _, a := range accs {
		for _, k := range sortedKinds(a.tokenWatches) {
			if emits[k] {
				return triggerOrderWhy{reason: TriggerOrderReasonReactor, event: k, token: true}
			}
		}
	}
	if g.writtenObjectsRippleLocked(written) {
		w := triggerOrderWhy{reason: TriggerOrderReasonRipple}
		// Name the first written object the records mention, if one alone does.
		for _, id := range sortedIDs(written) {
			if g.writtenObjectsRippleLocked(map[uuid.UUID]bool{id: true}) {
				w.card = g.whyCardNameLocked(id)
				break
			}
		}
		return w
	}
	// Who might leave the game while the batch resolves (CR 104.3b and
	// 104.3c, state-based losses checked between the resolutions): a player
	// the batch's life loss could bring to 0, or one it could make draw
	// from an empty library.
	eliminable := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if n := loss[p.ID]; n > 0 && p.Life-n <= 0 {
			eliminable[p.ID] = true
		}
		if n := draws[p.ID]; n > 0 && (p.Library == nil || n > len(p.Library.Cards)) {
			eliminable[p.ID] = true
		}
	}
	// A controller who can leave the game takes the rest of the batch
	// with them (CR 800.4a), so where they leave in it shows.
	for _, a := range accs {
		if eliminable[a.controller] {
			return triggerOrderWhy{reason: TriggerOrderReasonElimination, player: a.controller}
		}
	}
	// The board's layer statics, read once and only for a batch with a
	// characteristic read in it.
	var st *layerStatics
	statics := func() layerStatics {
		if st == nil {
			v := g.layerStaticsLocked()
			st = &v
		}
		return *st
	}
	for i := range accs {
		for j := i + 1; j < len(accs); j++ {
			if w := g.accessesConflictLocked(&accs[i], &accs[j], eliminable, statics); w.reason != "" {
				return w
			}
			if w := g.accessesConflictLocked(&accs[j], &accs[i], eliminable, statics); w.reason != "" {
				return w
			}
		}
	}
	return triggerOrderWhy{}
}

// sortedKinds lists a set of event kinds in a fixed order.
func sortedKinds(set map[EventKind]bool) []EventKind {
	out := make([]EventKind, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// accessesConflictLocked reports whether a's writes can change what b
// reads or writes, as the reason they conflict (with the facts its
// sentence names), or the zero value when they don't. Called both ways
// round for each pair.
func (g *Game) accessesConflictLocked(a, b *triggerAccess, eliminable map[uuid.UUID]bool, statics func() layerStatics) triggerOrderWhy {
	if a.rules {
		return triggerOrderWhy{reason: TriggerOrderReasonRules, card: g.whyCardNameLocked(a.source)}
	}
	if a.pumpRules && (b.charReads || b.traitReads || b.boardReads) {
		return triggerOrderWhy{reason: TriggerOrderReasonRules, card: g.whyCardNameLocked(a.source)}
	}
	// b reads a characteristic through the layers, and a changes
	// something a static on the board could read to compute it: a life
	// total for a static that declares it reads one, anything else for
	// any static that can change power or toughness (#3013).
	if card, conflict := characteristicReadConflict(a, b, statics); conflict {
		return triggerOrderWhy{reason: TriggerOrderReasonCharacteristic, card: card, life: accessLifeDirection(a)}
	}
	if a.monarch && b.monarch {
		return triggerOrderWhy{reason: TriggerOrderReasonMonarch}
	}
	// b's intervening "if" reads what a changes (#3033).
	if a.monarch && b.monarchRead {
		return triggerOrderWhy{reason: TriggerOrderReasonCondition, card: g.whyCardNameLocked(b.source)}
	}
	if b.boardReads && boardConditionSees(a, statics) {
		return triggerOrderWhy{reason: TriggerOrderReasonCondition, card: g.whyCardNameLocked(b.source)}
	}
	// A tie on one object asks, whatever the two do to it.
	for _, id := range sortedIDs(a.writes) {
		if b.writes[id] || b.reads[id] {
			return triggerOrderWhy{reason: TriggerOrderReasonObjectTie, card: g.whyCardNameLocked(id), both: b.writes[id]}
		}
	}
	for p, dir := range a.life {
		other := b.life[p]
		// Gains commute with gains and losses with losses. A gain
		// and a loss commute too, unless the player could reach 0 in
		// between and lose (CR 704.5a).
		if other != 0 && other != dir && eliminable[p] {
			return triggerOrderWhy{reason: TriggerOrderReasonElimination, player: p}
		}
	}
	// Two draws commute: the same cards reach the hand whichever item
	// takes them. A draw that empties the library is the elimination
	// rule's, and a "whenever you draw" watcher the reactor rule's.
	for p := range a.cards {
		if b.cards[p] && (a.beyondDraw[p] || b.beyondDraw[p]) {
			return triggerOrderWhy{reason: TriggerOrderReasonLibrary}
		}
	}
	for _, p := range sortedIDs(eliminable) {
		if a.life[p]&lifeDown == 0 && a.draws[p] == 0 {
			continue
		}
		// a can take p out of the game: then whatever b does to p, or
		// to anything p owns or controls, depends on whether b comes
		// first. Another loss of life is the same either way.
		// A player who leaves takes their permanents with them (CR
		// 800.4a), which a board condition sees.
		if b.life[p]&lifeUp != 0 || b.cards[p] || b.playerReads[p] || b.boardReads {
			return triggerOrderWhy{reason: TriggerOrderReasonElimination, player: p}
		}
		if g.touchesObjectsOfLocked(b, p) {
			return triggerOrderWhy{reason: TriggerOrderReasonElimination, player: p}
		}
	}
	return triggerOrderWhy{}
}

// boardConditionSees reports whether what a does can change a board
// condition (FootprintCondBoard) another item reads: a permanent it
// changes or pumps, a permanent or token it puts onto or takes off the
// battlefield, a counter it puts on one, the monarch or a player's
// counters (a static could read either without declaring it), and a
// life total, a hand or a library only when a static on the board
// declares that it reads one (#3033). The layer engine's contract
// (layerVersionBump) is that a static's output reads nothing else.
func boardConditionSees(a *triggerAccess, statics func() layerStatics) bool {
	if len(a.writes) > 0 || len(a.pumped) > 0 || a.monarch || a.rules || a.pumpRules {
		return true
	}
	for _, k := range []EventKind{EventTokenCreated, EventETB, EventZoneMove, EventLTB, EventCounterPlaced, EventPlayerCounterPlaced} {
		if a.emits[k] {
			return true
		}
	}
	if len(a.life) > 0 && statics().anyLife {
		return true
	}
	return len(a.cards) > 0 && statics().anyCards
}

// addConditionReadLocked adds what an intervening "if" reads, or reports
// that the check cannot say (#3033). Each condition is argued here:
//
//   - A tally that only rises this turn (life gained, life lost, a
//     creature died) and a designation that is never lost (the city's
//     blessing) cannot be made false: once it holds, it holds until the
//     batch has resolved, since a turn cannot end with the batch still on
//     the stack. So a condition that holds now reads nothing the batch
//     can change. One that does not hold yet could be made true by
//     another item, so the step is refused.
//   - "You cast no spells this turn" can only become false, by a cast,
//     and no step in the vocabulary casts. The same rule applies: it
//     reads nothing while it holds, and is refused once it does not.
//   - The monarch: an item that makes a player the monarch can change it
//     (monarchRead).
//   - The source: a read of the item's source object, as a counter on
//     it is.
//   - The board: boardReads, which boardConditionSees conflicts with.
//
// Caller must hold g.mu.
func (g *Game) addConditionReadLocked(a *triggerAccess, t *StackItem, s FootprintStep) bool {
	you := t.Controller
	tally := g.TurnTallyFor(you)
	atLeast := max(s.N, 1)
	switch s.Cond {
	case FootprintCondLifeGained:
		return tally.LifeGained >= atLeast
	case FootprintCondLifeLost:
		return tally.LifeLost >= atLeast
	case FootprintCondCreatureDied:
		return tally.CreaturesDied > 0
	case FootprintCondCitysBlessing:
		return g.CitysBlessingForEffect(you)
	case FootprintCondNoSpellsCast:
		return g.CastTallyFor(you).Total == 0
	case FootprintCondMonarch:
		a.monarchRead = true
	case FootprintCondSource:
		a.reads[t.SourceCardID] = true
	case FootprintCondBoard:
		a.boardReads = true
	default:
		return false
	}
	return true
}

// accessLifeDirection is the way an item moves life totals, folded across players.
func accessLifeDirection(a *triggerAccess) lifeDirection {
	var d lifeDirection
	for _, dir := range a.life {
		d |= dir
	}
	return d
}

// characteristicReadConflict reports whether b reads a characteristic
// through the layers and a changes something a static on the board
// could read to compute it: a life total for a static that declares it
// reads one, anything else for any static that can change power or
// toughness (#3013). It is the whole test for TriggerOrderReasonCharacteristic,
// and it names the card behind the static it found.
func characteristicReadConflict(a, b *triggerAccess, statics func() layerStatics) (string, bool) {
	if !b.charReads || !a.writesAnything() {
		return "", false
	}
	st := statics()
	if st.life {
		return st.lifeCard, true
	}
	if st.pt && !a.writesOnlyLife() {
		return st.ptCard, true
	}
	return "", false
}

// identicalBatchOrderShowsLocked is #3021: the sub-PR 8 skip (every
// item has one source and one label) argues the items are
// interchangeable, and that holds only while they read nothing that
// differs between them. An item that reads its triggering subject
// (GainLifeEqualToSubject: Righteous Valkyrie, Trostani) reads a
// DIFFERENT subject in each copy, through the layers, so one copy's
// gain can change what the next reads (Valkyrie's +7 line gives +2/+2).
//
// This reports true only for that case: the mode is "when it matters",
// the batch still has an unanswered item, every item is identical by
// the sub-PR 8 test and is one the footprint check can describe, and
// one copy's writes could change what another reads through a static
// on the board (characteristicReadConflict). Everything else the old
// skip let through (two Bident draws, two counters on one permanent,
// an item the check cannot describe) stays skipped: this adds no
// prompt that the footprint check was not already asked for, and asks
// for no reason but the one that is true.
//
// Caller must hold g.mu.
func (g *Game) identicalBatchOrderShowsLocked(items []*StackItem, mode TriggerOrderMode) (triggerOrderWhy, bool) {
	if len(items) < 2 || mode == TriggerOrderAlways || mode == TriggerOrderNever {
		return triggerOrderWhy{}, false
	}
	anyOpen := false
	for _, t := range items {
		if !t.Ordered {
			anyOpen = true
		}
	}
	if !anyOpen || !identicalItems(items) {
		return triggerOrderWhy{}, false
	}
	accs := make([]triggerAccess, len(items))
	reads := false
	for i, t := range items {
		a, ok := g.triggerAccessLocked(t)
		if !ok {
			return triggerOrderWhy{}, false
		}
		accs[i] = a
		reads = reads || a.charReads
	}
	if !reads {
		return triggerOrderWhy{}, false
	}
	var st *layerStatics
	statics := func() layerStatics {
		if st == nil {
			v := g.layerStaticsLocked()
			st = &v
		}
		return *st
	}
	for i := range accs {
		for j := range accs {
			if i == j {
				continue
			}
			if card, conflict := characteristicReadConflict(&accs[i], &accs[j], statics); conflict {
				return triggerOrderWhy{
					reason: TriggerOrderReasonCharacteristic, card: card, life: accessLifeDirection(&accs[i]), identical: true,
				}, true
			}
		}
	}
	return triggerOrderWhy{}, false
}

// touchesObjectsOfLocked reports whether the access reads or writes an
// object player p owns or controls.
func (g *Game) touchesObjectsOfLocked(a *triggerAccess, p uuid.UUID) bool {
	for _, m := range []map[uuid.UUID]bool{a.reads, a.writes} {
		for id := range m {
			c, ok := g.LookupCardForEffect(id)
			if !ok || c.Owner == p || c.Controller == p {
				return true
			}
		}
	}
	return false
}

// triggerAccessLocked resolves one item's footprint, or reports that
// the item is one the check cannot describe.
func (g *Game) triggerAccessLocked(t *StackItem) (triggerAccess, bool) {
	if t == nil || t.Kind != StackItemTriggered {
		return triggerAccess{}, false
	}
	// #2494: an optional trigger its chooser has a standing "No" for
	// resolves doing nothing (CR 603.5), so it touches nothing whatever
	// its place in the batch. Every other optional item is undescribed
	// below: its "you may" is decided on the board it resolves on.
	if g.optionalStandingDeclineLocked(t) {
		return newTriggerAccess(t.Controller), true
	}
	if steps, ok := keyedBodyFootprints[t.Body]; ok && t.Body != CatalogTriggeredBodyKey {
		return g.keyedTriggerAccessLocked(t, steps)
	}
	if t.Body != CatalogTriggeredBodyKey || t.Params.Ability == nil {
		return triggerAccess{}, false
	}
	row, _, outcome := resolveTriggeredAbilityRef(*t.Params.Ability)
	if outcome != abilityRefMatched || len(row.Footprint) == 0 {
		return triggerAccess{}, false
	}
	// Unknown: a "you may", a mode clause or chosen modes, a clause
	// built from the trigger, or anything recorded on the item for it
	// alone.
	if row.OptionalPrompt != nil || row.Modes != nil ||
		row.TargetsFrom != nil || row.TargetsFromReadsBoard || len(t.Modes) > 0 {
		return triggerAccess{}, false
	}
	// A Build fill-in (owner answer to ADR 0018's #2884 question,
	// 2026-10-09) is accepted only when the item it built is exactly
	// what the footprint describes: the effect is still the row's
	// declared one, and the item carries nothing beyond the fields the
	// engine stamps and the footprint reads.
	if row.Build != nil && !builtItemIsPlain(t, row) {
		return triggerAccess{}, false
	}
	params := t.Params
	params.Ability = nil
	// An X the engine stamped from its own cast record (XFromCast) is a
	// constant of the item, fixed before the batch exists (#3033).
	if !reflect.ValueOf(params).IsZero() || len(t.Payload) > 0 || (t.XValue != 0 && !row.XFromCast) || len(t.Distribution) > 0 {
		return triggerAccess{}, false
	}
	a := newTriggerAccess(t.Controller)
	var targetCards []uuid.UUID
	for _, ref := range t.Targets {
		switch ref.Kind {
		case TargetCard:
			a.reads[ref.ID] = true
			targetCards = append(targetCards, ref.ID)
		case TargetPlayer:
			a.playerReads[ref.ID] = true
		default:
			return triggerAccess{}, false
		}
	}
	if len(targetCards) > 0 || len(a.playerReads) > 0 {
		// CR 608.2b: a target's legality is read through the layers.
		a.traitReads = true
	}
	for _, s := range row.Footprint {
		if !g.addFootprintStepLocked(&a, t, s, targetCards) {
			return triggerAccess{}, false
		}
	}
	g.markRulesLocked(&a)
	return a, true
}

// keyedTriggerAccessLocked resolves an engine keyword trigger's item
// (prowess) against its body's declared footprint. The item must be
// exactly what the keyword's Build makes: no targets, modes, Params,
// payload, X or division, and no field beyond the ones a plain built
// item may set.
func (g *Game) keyedTriggerAccessLocked(t *StackItem, steps []FootprintStep) (triggerAccess, bool) {
	if len(t.Targets) > 0 || len(t.Modes) > 0 || !reflect.ValueOf(t.Params).IsZero() ||
		len(t.Payload) > 0 || t.XValue != 0 || len(t.Distribution) > 0 ||
		!builtItemIsPlain(t, TriggeredAbility{}) {
		return triggerAccess{}, false
	}
	a := newTriggerAccess(t.Controller)
	for _, s := range steps {
		if !g.addFootprintStepLocked(&a, t, s, nil) {
			return triggerAccess{}, false
		}
	}
	g.markRulesLocked(&a)
	return a, true
}

// markRulesLocked sets rules for an item that writes an object whose
// abilities shape the rules the other items resolve under, and
// pumpRules for one whose only change to such an object is a pump. It
// then counts every pumped object as written.
func (g *Game) markRulesLocked(a *triggerAccess) {
	for id := range a.writes {
		if !g.objectInertLocked(id) {
			a.rules = true
		}
	}
	for id := range a.pumped {
		if !a.writes[id] && !g.objectInertLocked(id) {
			a.pumpRules = true
		}
	}
	for id := range a.pumped {
		a.writes[id] = true
	}
}

// builtItemPlainFields are the StackItem fields a built item may set and
// still be read by its footprint alone: the engine's own stamps (identity,
// kind, ordering, the triggering event, the source object, the catalog
// row), the controller and owner the footprint resolves "you" against,
// the label, and the targets of a declared target clause. Params are
// checked separately (only the row stamp may be set). Every other field
// a Build sets makes the item one the check cannot describe.
var builtItemPlainFields = map[string]bool{
	"ID": true, "Kind": true, "Controller": true, "BaseController": true, "Owner": true,
	"SourceCardID": true, "SourceEpoch": true, "SourceObject": true, "Label": true,
	"DoubledBy": true, "DoubledByName": true, "Targets": true, "Trigger": true,
	"Seq": true, "Effect": true, "Body": true, "Params": true, "Ordered": true,
	"TargetsAnnouncePending": true, "targetSpec": true,
	// #1511's ordering mark: it says how a batch may be ordered and
	// changes nothing the item does (#3013, prowess).
	"Commutes": true,
}

// builtItemIsPlain reports whether an item a row's Build made sets only
// builtItemPlainFields, and targets only through the row's declared
// target clause.
func builtItemIsPlain(t *StackItem, row TriggeredAbility) bool {
	if len(t.Targets) > 0 && row.Targets == nil {
		return false
	}
	v := reflect.ValueOf(t).Elem()
	ty := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if !builtItemPlainFields[ty.Field(i).Name] && !v.Field(i).IsZero() {
			return false
		}
	}
	return true
}

// addFootprintStepLocked adds one step's reads, writes and events.
func (g *Game) addFootprintStepLocked(a *triggerAccess, t *StackItem, s FootprintStep, targets []uuid.UUID) bool {
	you := t.Controller
	switch s.Kind {
	case FootprintGainLife:
		if s.AmountFromSubject {
			if t.Trigger == nil || t.Trigger.Object == nil {
				return false
			}
			subject := t.Trigger.Object.ID
			if z := g.findCardZoneLocked(subject); z != nil && z.Kind == ZoneBattlefield {
				a.reads[subject] = true
				a.charReads = true
			}
		}
		a.life[you] |= lifeUp
		a.emit(EventChangeLife)
	case FootprintDamageEachOpponent:
		if !g.footprintDamageSourceLocked(a, t.SourceCardID) {
			return false
		}
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == you {
				continue
			}
			a.life[p.ID] |= lifeDown
			a.lifeLoss[p.ID] += s.N
		}
		a.emit(EventDealDamage, EventChangeLife)
	case FootprintDamageController:
		if !g.footprintDamageSourceLocked(a, t.SourceCardID) {
			return false
		}
		a.life[you] |= lifeDown
		a.lifeLoss[you] += s.N
		a.emit(EventDealDamage, EventChangeLife)
	case FootprintTargetPlayerLosesLife:
		// The first player target, the one the step acts on. An item
		// with none declares nothing to lose, and is refused rather than
		// read as a no-op.
		var target uuid.UUID
		for _, ref := range t.Targets {
			if ref.Kind == TargetPlayer {
				target = ref.ID
				break
			}
		}
		if target == uuid.Nil {
			return false
		}
		a.playerReads[target] = true
		a.life[target] |= lifeDown
		a.lifeLoss[target] += s.N
		a.emit(EventChangeLife)
	case FootprintLoseLifeEachOpponent:
		// The source is passed on as the loss's attribution and nothing
		// else reads it, so the step reads no object.
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == you {
				continue
			}
			a.life[p.ID] |= lifeDown
			a.lifeLoss[p.ID] += s.N
		}
		a.emit(EventChangeLife)
	case FootprintCounterOnSource:
		a.reads[t.SourceCardID] = true
		if z := g.findCardZoneLocked(t.SourceCardID); z != nil && z.Kind == ZoneBattlefield {
			a.writes[t.SourceCardID] = true
		}
		a.emit(EventCounterPlaced)
	case FootprintExileTargets:
		for _, id := range targets {
			a.writes[id] = true
		}
		a.emit(EventZoneMove, EventLTB, EventUnattach)
	case FootprintDraw:
		a.cards[you] = true
		a.draws[you] += s.N
		a.emit(EventDrawCard)
	case FootprintMill:
		a.cards[you] = true
		a.beyondDraw[you] = true
		a.emit(EventMill, EventZoneMove)
	case FootprintScry:
		a.cards[you] = true
		a.beyondDraw[you] = true
		a.emit(EventScry)
	case FootprintSurveil:
		a.cards[you] = true
		a.beyondDraw[you] = true
		// The graveyard leg emits EventMill per card, as an ordinary
		// mill does (ResolveSurveil): a "card put into your graveyard
		// from your library" watcher sees it. Found by the footprint
		// conformance test (#3013).
		a.emit(EventSurveil, EventMill, EventZoneMove)
	case FootprintEnergy:
		a.emit(EventPlayerCounterPlaced)
	case FootprintCreateToken:
		if !tokenTemplateInert(s.TokenKey, s.TokenManual) {
			return false
		}
		for _, k := range tokenKeywordWatches(s.TokenKey, s.TokenKeywords) {
			a.tokenWatches[k] = true
		}
		a.emit(EventTokenCreated, EventETB, EventZoneMove)
	case FootprintMonarch:
		a.monarch = true
		a.emit(EventMonarchChanged)
	case FootprintCondition:
		return g.addConditionReadLocked(a, t, s)
	case FootprintPumpSource:
		// resolveProwess: nothing for a source that left, or left and
		// came back as a new object (CR 400.7); otherwise a layer-7c
		// record pinned to it. Registering a record emits no event.
		a.reads[t.SourceCardID] = true
		if z := g.findCardZoneLocked(t.SourceCardID); z != nil && z.Kind == ZoneBattlefield {
			a.pumped[t.SourceCardID] = true
		}
	default:
		return false
	}
	return true
}

// footprintDamageSourceLocked is the read a damage step makes of its
// source, the item's source card, or false when the check cannot say
// what that damage does. It asks the damage tail the step will build
// (effectDamageTailLocked, as dealDamageToPlayerLocked does), so the two
// cannot disagree about the source's traits: the permanent's while it is
// on the battlefield, its last-known record once it has left (CR 608.2h),
// and none once the card has moved on and no record matches. The last
// two are the aristocrats case: Syr Konrad dies in the wipe whose deaths
// he triggers on, and as a commander goes on to the command zone. A
// departed source's traits cannot change, but another item that moves
// the card again (back to the battlefield) changes which object deals the
// damage, so the source is a read either way. Damage that does more than
// cost life — lifelink (CR 702.15b), infect, wither, toxic — is not on
// the list.
//
// Caller must hold g.mu.
func (g *Game) footprintDamageSourceLocked(a *triggerAccess, source uuid.UUID) bool {
	tail := g.effectDamageTailLocked(damageTailPlayer, source, nil)
	if tail.lifelinkTo != uuid.Nil || tail.result != (DamageResultSource{}) {
		return false
	}
	a.reads[source] = true
	a.traitReads = true
	return true
}

// FootprintDeclaration is what the independence check makes of one
// FootprintStep resolved against an item and the board: the objects it
// reads and writes, the players whose life it moves up or down, whose
// library, hand or graveyard it touches, whether it makes its controller
// the monarch, and the event kinds it causes. It exists for the
// conformance test (cards/effects/footprint_conformance_test.go), which
// resolves each primitive and checks that it changes and emits nothing
// outside this.
type FootprintDeclaration struct {
	Reads, Writes    map[uuid.UUID]bool
	LifeUp, LifeDown map[uuid.UUID]bool
	Cards            map[uuid.UUID]bool
	Monarch          bool
	Emits            map[EventKind]bool
}

// FootprintDeclarationForTest describes one step the way
// addFootprintStepLocked does for the check, with the item's card
// targets as its targets. False when the check refuses the step on this
// board.
//
// Caller must hold g.mu.
func (g *Game) FootprintDeclarationForTest(item *StackItem, s FootprintStep) (FootprintDeclaration, bool) {
	a := newTriggerAccess(item.Controller)
	var targets []uuid.UUID
	for _, ref := range item.Targets {
		if ref.Kind == TargetCard {
			a.reads[ref.ID] = true
			targets = append(targets, ref.ID)
		}
	}
	if !g.addFootprintStepLocked(&a, item, s, targets) {
		return FootprintDeclaration{}, false
	}
	for id := range a.pumped {
		a.writes[id] = true
	}
	d := FootprintDeclaration{
		Reads: a.reads, Writes: a.writes, Cards: a.cards, Monarch: a.monarch, Emits: a.emits,
		LifeUp: map[uuid.UUID]bool{}, LifeDown: map[uuid.UUID]bool{},
	}
	for p, dir := range a.life {
		if dir&lifeUp != 0 {
			d.LifeUp[p] = true
		}
		if dir&lifeDown != 0 {
			d.LifeDown[p] = true
		}
	}
	return d, true
}

// tokenKeywordWatches is every event kind the keyword triggers a token
// made from the template would have watch: its own keywords and its
// catalog entry's printed ones, through the walk the harvester uses
// (keywordTriggersFor).
func tokenKeywordWatches(key string, keywords []string) []EventKind {
	probe := Card{TokenKey: key, Keywords: keywords}
	var out []EventKind
	for _, t := range keywordTriggersFor(&probe) {
		out = append(out, t.Watches...)
	}
	return out
}

// tokenTemplateInert reports whether a token made from tmpl brings
// nothing onto the battlefield that reads or changes the game: no
// triggered ability (its own "when this enters" would react to its own
// entry), no static, replacement or rules-bearing slot, and no printed
// text the engine leaves to the table.
func tokenTemplateInert(key string, manual bool) bool {
	if manual {
		return false
	}
	d := catalogDef(key)
	if d == nil {
		return true
	}
	if len(d.Triggered) > 0 {
		return false
	}
	global, own := defBearsRules(d)
	return !global && !own
}

// objectInertLocked reports whether changing the object (its zone or
// its counters) changes nothing but the object itself: its text is the
// engine's, it has no static, replacement or rules-bearing slot, it is
// attached to nothing and nothing is attached to it.
func (g *Game) objectInertLocked(id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	if !ok || Unimplemented(c) || c.AttachedTo.ID != uuid.Nil {
		return false
	}
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].AttachedTo.ID == id {
				return false
			}
		}
	}
	d := catalogDef(catalogAbilityKeyOf(&c))
	if d == nil {
		return true
	}
	global, own := defBearsRules(d)
	return !global && !own
}

// triggerOrderBoardOpaqueLocked reports a board the check cannot read
// at all: a permanent or emblem with a rules-bearing slot (a "can't
// gain life", a "damage can't be prevented", a player's hexproof), a
// permanent whose text the engine leaves to the table, a state trigger
// (CR 603.8: it reads the game state itself), a player static, or a
// continuing effect that is more than a fixed change to the
// characteristics of the objects it locked in (CR 611.2c).
//
// Caller must hold g.mu.
func (g *Game) triggerOrderBoardOpaqueLocked() (card string, opaque bool) {
	zones := []*Zone{g.Battlefield}
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		if len(p.Statics) > 0 {
			return "", true
		}
		zones = append(zones, p.Emblems)
	}
	for _, z := range zones {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			if Unimplemented(*c) {
				return whyCardName(*c), true
			}
			if d := catalogDef(catalogAbilityKeyOf(c)); d != nil {
				if global, _ := defBearsRules(d); global {
					return whyCardName(*c), true
				}
			}
			for _, t := range triggersOf(c) {
				if t.State != "" {
					return whyCardName(*c), true
				}
			}
		}
	}
	for i := range g.ScopedEffects {
		se := &g.ScopedEffects[i]
		if se.Scope != ScopeNone {
			return "", true
		}
		for _, m := range se.Mods {
			if !characteristicModKinds[m.Kind] {
				return "", true
			}
		}
	}
	return "", false
}

// characteristicModKinds are the continuing-effect mods that only change
// the characteristics or the combat rules of the objects a record locked
// in (layers 1 to 7, and the layer-6 attack and block rules). Whatever
// they do to an object, they do it in every order. Every other mod (a
// prevention or replacement, a "can't", a doubling) changes what an
// event does, and makes the board opaque to the check.
var characteristicModKinds = map[ModKind]bool{
	ModSetController: true, ModAddTypes: true, ModRemoveTypes: true,
	ModAddSubtypes: true, ModAllCreatureTypes: true, ModSetColors: true, ModAddColors: true,
	ModAddKeywords: true, ModRemoveKeywords: true, ModLoseAllAbilities: true,
	ModLoseOwnAbility: true, ModAddRestrictions: true, ModSetBasePower: true,
	ModSetBaseToughness: true, ModModifyPT: true, ModAddAttackRequirement: true, ModAttackPlayerIfAble: true, ModGoad: true,
	ModAddBlockRequirement: true, ModAddBlockCapacity: true, ModBlockAnyNumber: true,
	ModGrantAbilities: true, ModBorrowAbilities: true, ModCantAttackUnlessDefenderControls: true, ModCantAttackPlayer: true, ModCantAttackOwner: true,
	ModSetBasicLandTypes: true, ModLoseLandTypes: true, ModCantHaveKeywords: true,
	ModBecomeCopy: true, ModExchangeText: true, ModChangeText: true, ModCantBeBlockedExceptBy: true,
	ModLimitBlockersPerDefender: true, ModCantBeBlockedByPlayer: true,
	ModCantBeBlockedByPower: true, ModSetCreatureTypes: true,
}

// knownBuiltinReplacementLabels are the engine's own replacement effects
// the check sets aside, each for a reason that needs no board read:
// CR 903.9's commander move asks the commander's owner whatever the
// order; a regeneration shield replaces destruction, which no step
// does; protection's damage prevention and a locked life total need a
// player with protection or a lock, which is a player static or a
// rules-bearing slot and makes the board opaque already.
var knownBuiltinReplacementLabels = map[string]bool{
	commanderZoneReplacement.Label:            true,
	regenerationShieldReplacement.Label:       true,
	protectionPreventsDamageReplacement.Label: true,
	lifeTotalCantChangeReplacement.Label:      true,
	// The entry-counters grant exists only while a player static does,
	// and any player static makes the board opaque already.
	playerEntryCountersReplacement.Label: true,
}

// somethingReactsLocked reports whether anything in the game watches
// one of the event kinds the batch causes: a triggered ability on a
// permanent or an emblem (whatever its condition — the check does not
// guess whether it would fire), one that functions from another zone, a
// delayed trigger or an "until" return, a replacement effect, or speed's
// inherent trigger. Its result would go on the stack between the
// batch's items, so where it lands depends on the order.
//
// Caller must hold g.mu.
func (g *Game) reactorLocked(emits map[EventKind]bool) (card string, kind EventKind, reacts bool) {
	watches := func(kinds []EventKind) (EventKind, bool) {
		for _, k := range kinds {
			if emits[k] {
				return k, true
			}
		}
		return "", false
	}
	zones := []*Zone{g.Battlefield}
	for _, p := range g.Seats {
		if p != nil {
			zones = append(zones, p.Emblems)
		}
	}
	for _, z := range zones {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			for _, t := range triggersOf(c) {
				if k, ok := watches(t.Watches); ok {
					return whyCardName(*c), k, true
				}
			}
			if CatalogReplacements == nil {
				continue
			}
			for _, r := range CatalogReplacements(catalogAbilityKeyOf(c)) {
				// An entry-only replacement ("This land enters tapped")
				// changes only its own object's entry, which a permanent
				// already in play has made. No described step puts that
				// object onto the battlefield again, so it reacts to
				// nothing the batch does. Every tapland on the board
				// made each token or zone-moving batch ask until
				// 2026-10-14.
				if r.EntryOnly && z == g.Battlefield {
					continue
				}
				k, ok := watches(r.Watches)
				if len(r.Watches) == 0 || ok {
					return whyCardName(*c), k, true
				}
			}
		}
	}
	for k := range emits {
		for _, zk := range triggerZones.zonesFor(k) {
			for _, z := range g.triggerZonesOfKindLocked(zk) {
				for i := range z.Cards {
					if triggerZones.declares(CatalogKey(z.Cards[i])) {
						return whyCardName(z.Cards[i]), k, true
					}
				}
			}
		}
	}
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil || CatalogReplacements == nil {
			continue
		}
		for i := range p.Graveyard.Cards {
			for _, r := range CatalogReplacements(catalogAbilityKeyOf(&p.Graveyard.Cards[i])) {
				k, ok := watches(r.Watches)
				if r.FromGraveyard && (len(r.Watches) == 0 || ok) {
					return whyCardName(p.Graveyard.Cards[i]), k, true
				}
			}
		}
	}
	for _, dt := range g.DelayedTriggers {
		if dt == nil {
			continue
		}
		if k, ok := watches(dt.On); ok {
			return "", k, true
		}
	}
	for i := range g.BuiltinReplacements {
		r := &g.BuiltinReplacements[i]
		k, ok := watches(r.Watches)
		if r.Label == shieldCounterLabel {
			// CR 122.1c's effects apply only to a permanent with a
			// shield counter on it (#2000, ADR 0173): set aside on a
			// board with none.
			if ok && g.anyShieldCounterLocked() {
				return "", k, true
			}
			continue
		}
		if !knownBuiltinReplacementLabels[r.Label] && (len(r.Watches) == 0 || ok) {
			return "", k, true
		}
	}
	// CR 702.179d: the active player's speed goes up when an opponent
	// loses life during their turn — an inherent trigger with no card.
	if emits[EventChangeLife] || emits[EventDealDamage] {
		if a := g.Turn.ActiveSeat; a >= 0 && a < len(g.Seats) && g.Seats[a] != nil && g.Seats[a].Speed > 0 {
			if emits[EventChangeLife] {
				return "", EventChangeLife, true
			}
			return "", EventDealDamage, true
		}
	}
	return "", "", false
}

// writtenObjectsRippleLocked reports whether changing one of the
// written objects reaches past it: a delayed trigger, an "until"
// return or a continuing effect's duration that names it, or another
// card that records it (exiled with it, paired with it, attached to
// it). Read by walking those records for the object's ID, which is
// blunt on purpose: a record the check has never heard of still counts.
//
// Caller must hold g.mu.
func (g *Game) writtenObjectsRippleLocked(written map[uuid.UUID]bool) bool {
	if len(written) == 0 {
		return false
	}
	seen := map[uintptr]bool{}
	for _, dt := range g.DelayedTriggers {
		if dt != nil && mentionsAnyID(reflect.ValueOf(dt), written, seen) {
			return true
		}
	}
	for i := range g.ScopedEffects {
		if mentionsAnyID(reflect.ValueOf(g.ScopedEffects[i].Duration), written, seen) {
			return true
		}
	}
	zones := []*Zone{g.Battlefield, g.Exile}
	for _, p := range g.Seats {
		if p != nil {
			zones = append(zones, p.Graveyard, p.Command)
		}
	}
	for _, z := range zones {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			if written[c.InstanceID] {
				continue
			}
			if mentionsAnyID(reflect.ValueOf(c), written, seen) {
				return true
			}
		}
	}
	return false
}

var independenceUUIDType = reflect.TypeOf(uuid.UUID{})

// mentionsAnyID walks v for a uuid.UUID in ids. Functions and channels
// are not walked; each pointer is walked once.
func mentionsAnyID(v reflect.Value, ids map[uuid.UUID]bool, seen map[uintptr]bool) bool {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return false
		}
		if p := v.Pointer(); seen[p] {
			return false
		} else {
			seen[p] = true
		}
		return mentionsAnyID(v.Elem(), ids, seen)
	case reflect.Interface:
		if v.IsNil() {
			return false
		}
		return mentionsAnyID(v.Elem(), ids, seen)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if mentionsAnyID(v.Field(i), ids, seen) {
				return true
			}
		}
	case reflect.Array:
		if v.Type() == independenceUUIDType {
			var id uuid.UUID
			for i := range id {
				id[i] = byte(v.Index(i).Uint())
			}
			return ids[id]
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return false
		}
		for i := 0; i < v.Len(); i++ {
			if mentionsAnyID(v.Index(i), ids, seen) {
				return true
			}
		}
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return false
		}
		for i := 0; i < v.Len(); i++ {
			if mentionsAnyID(v.Index(i), ids, seen) {
				return true
			}
		}
	case reflect.Map:
		it := v.MapRange()
		for it.Next() {
			if mentionsAnyID(it.Key(), ids, seen) || mentionsAnyID(it.Value(), ids, seen) {
				return true
			}
		}
	}
	return false
}

// defFieldClass is how one CardDef slot bears on the independence
// check.
type defFieldClass uint8

const (
	// defFieldInert: the slot cannot change what a footprint step does
	// as it resolves (casting, costs, combat, mana, the untap step).
	defFieldInert defFieldClass = iota + 1
	// defFieldReactor: read by somethingReactsLocked (triggers,
	// replacement effects) rather than here.
	defFieldReactor
	// defFieldOwnRules: the slot changes the game while its permanent
	// is there (a static ability, a trigger doubler), so changing that
	// permanent changes the rules the other items resolve under.
	defFieldOwnRules
	// defFieldRules: a rule a footprint step obeys as it resolves
	// ("can't gain life", "damage can't be prevented", a player's
	// hexproof, "can't lose"). Its presence anywhere makes the board
	// opaque: its own condition could read what an item writes.
	defFieldRules
)

// cardDefFieldClass classifies every CardDef slot. A slot missing from
// the map counts as defFieldRules, and
// TestEveryCardDefSlotHasAnIndependenceClass fails until it is added.
var cardDefFieldClass = map[string]defFieldClass{
	"Resolve": defFieldInert, "AsEnters": defFieldInert, "AsEntersOthers": defFieldInert, "AsTransformsInto": defFieldInert, "AsAttached": defFieldInert,
	"StartingLoyalty": defFieldInert, "BattleDefense": defFieldInert,
	"TargetMode": defFieldInert, "Targets": defFieldInert, "Modes": defFieldInert,
	"Purpose": defFieldInert, "ManaAbilities": defFieldInert, "Activated": defFieldInert,
	"EntersWithCountersFromCast": defFieldInert, "PrintedKeywords": defFieldInert,
	"ManaTriggers": defFieldInert, "AdditionalCost": defFieldInert,
	"OptionalCosts": defFieldInert, "AlternativeCosts": defFieldInert,
	"TapCost": defFieldInert, "Delve": defFieldInert, "SpendOnly": defFieldInert,
	"SpendOnlySources": defFieldInert, "SpellsYouCastHaveDelve": defFieldInert, "SplitSecondGrants": defFieldInert, "GrantedOptionalCosts": defFieldInert, "GrantedTapCosts": defFieldInert,
	"CostModifiers": defFieldInert, "SelfCostModifiers": defFieldInert,
	"ExhaustPermissions": defFieldInert, "BoastLimits": defFieldInert,
	"AttackTaxes": defFieldInert, "TargetTaxes": defFieldInert, "BlockRules": defFieldInert, "AttackLimits": defFieldInert,
	"ExertOnAttack": defFieldInert, "CastableZones": defFieldInert, "CastableZonesIf": defFieldInert,
	"SpecialActions": defFieldInert, "SpecialActionGrants": defFieldInert,
	"UntapStep": defFieldInert, "UntapStepRestrictions": defFieldInert,
	"CounterRemovalLocks": defFieldInert, "UntapCaps": defFieldInert,
	"UntapOptOuts": defFieldInert, "DrawStep": defFieldInert,
	"CastCondition": defFieldInert, "CastConditionLabel": defFieldInert,
	"CastRestrictions": defFieldInert, "LandPlayRestrictions": defFieldInert,
	"EntryRestrictions":      defFieldInert,
	"ActivationRestrictions": defFieldInert, "ActivationTimings": defFieldInert,
	"CantBeCountered": defFieldInert, "CantBeCounteredIf": defFieldInert,
	"FlashWithCleanupSacrifice": defFieldInert, "CastTimeValues": defFieldInert,
	"SpellDamageCantBePrevented": defFieldInert, "SpellsCantBeCountered": defFieldInert,
	"HandSize": defFieldInert, "ManaPool": defFieldInert,
	"DamageStaysThroughCleanup": defFieldInert, "AnyColorSpend": defFieldInert, "LoyaltyTwiceEachTurn": defFieldInert, "ActivateAsThoughHaste": defFieldInert,
	// ADR 0146: an extra vote changes how a vote resolves while its
	// permanent is there.
	"ExtraVote": defFieldOwnRules,
	// #3140: CR 701.55c, the count a villainous choice is faced.
	"OpponentsFaceVillainousChoicesAgain": defFieldOwnRules,
	"LifeForMana":                         defFieldInert, "Emblem": defFieldInert, "TokenText": defFieldInert,
	"GrantText": defFieldInert, "MadnessLife": defFieldInert, "XMatters": defFieldInert, "XCeiling": defFieldInert,
	"WantsDistinctColors": defFieldInert, "WantsManaFrom": defFieldInert, "WantsColors": defFieldInert, "ReadsManaSpent": defFieldInert,
	"AdditionalLandPlays": defFieldInert, "CastPermissions": defFieldInert,
	"GatedCastPermissions": defFieldInert, "CastTimings": defFieldInert,
	"GrantedAlternativeCosts": defFieldInert, "OpeningHand": defFieldInert,
	"LibraryTopVisible": defFieldInert, "HandsRevealed": defFieldInert, "DiscardByOpponent": defFieldInert, "DiscardByOpponentForHand": defFieldInert,

	"Triggered": defFieldReactor, "Replacements": defFieldOwnRules,

	"Static": defFieldOwnRules, "TriggerDoublers": defFieldOwnRules,
	"TriggerSuppressors": defFieldOwnRules,

	"HexproofBypasses": defFieldRules, "WardSuppressions": defFieldRules,
	"TargetingRestrictions": defFieldRules, "PlayerKeywords": defFieldRules,
	"PlayerLifeTotalLocked": defFieldRules, "DamageCantBePrevented": defFieldRules,
	"CantGainLife": defFieldRules, "DamageAsThough": defFieldRules,
	"LegendRuleExemptions": defFieldRules, "ZeroLoyaltyExemptions": defFieldRules,
	"OpponentEffectProtections": defFieldRules, "GameEndGates": defFieldRules,
	"DealtDamageCantBeRegenerated": defFieldRules,
}

// defBearsRules reports whether a definition has a rules slot set
// (global: the board is opaque) or an own-rules slot set (own: changing
// this permanent changes the rules).
func defBearsRules(d *CardDef) (global, own bool) {
	v := reflect.ValueOf(d).Elem()
	ty := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := ty.Field(i)
		if !f.IsExported() || v.Field(i).IsZero() {
			continue
		}
		// The row the registry makes from the printed keywords applies
		// them to the object itself and reads nothing (#3013).
		if f.Name == "Static" && onlyPrintedKeywordRows(d.Static) {
			continue
		}
		switch class, ok := cardDefFieldClass[f.Name]; {
		case !ok || class == defFieldRules:
			global = true
		case class == defFieldOwnRules:
			own = true
		}
	}
	return global, own
}

// onlyPrintedKeywordRows reports whether every static is the row the
// registry makes from a card's printed keywords.
func onlyPrintedKeywordRows(statics []StaticAbility) bool {
	for i := range statics {
		if !statics[i].PrintedKeywordRow {
			return false
		}
	}
	return true
}
