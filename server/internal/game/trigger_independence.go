package game

import (
	"reflect"

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
	// FootprintGainLife: the item's controller gains N life.
	FootprintGainLife FootprintKind = "gain_life"
	// FootprintDamageEachOpponent: the item's source deals N damage to
	// each opponent of the item's controller.
	FootprintDamageEachOpponent FootprintKind = "damage_each_opponent"
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
)

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
}

// FootprintTokenOf describes a token template for a FootprintStep.
func FootprintTokenOf(tmpl Card) (key string, manual bool) {
	return catalogKeyOf(&tmpl), Unimplemented(tmpl)
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
	// Objects (cards and permanents), by instance ID.
	reads, writes map[uuid.UUID]bool
	// Players.
	life        map[uuid.UUID]lifeDirection
	lifeLoss    map[uuid.UUID]int
	cards       map[uuid.UUID]bool // library, hand and graveyard order
	draws       map[uuid.UUID]int
	playerReads map[uuid.UUID]bool // a player target
	// monarch is the one game-wide designation a step writes.
	monarch bool
	// rules is set when the item changes an object whose abilities
	// shape the rules every other item resolves under (a static
	// ability, a replacement effect, an attachment, a record that
	// names it). Everything reads the rules, so it conflicts with all.
	rules bool
	emits map[EventKind]bool
}

func newTriggerAccess(controller uuid.UUID) triggerAccess {
	return triggerAccess{
		controller:  controller,
		reads:       map[uuid.UUID]bool{},
		writes:      map[uuid.UUID]bool{},
		life:        map[uuid.UUID]lifeDirection{},
		lifeLoss:    map[uuid.UUID]int{},
		cards:       map[uuid.UUID]bool{},
		draws:       map[uuid.UUID]int{},
		playerReads: map[uuid.UUID]bool{},
		emits:       map[EventKind]bool{},
	}
}

func (a *triggerAccess) emit(kinds ...EventKind) {
	for _, k := range kinds {
		a.emits[k] = true
	}
}

// seatNeedsTriggerOrderLocked is seatNeedsTriggerOrder with the board
// in hand: in "when it matters" mode (and for a mode this binary does
// not know, which behaves as the default) a batch the pure skips would
// ask about is also skipped when its items are pairwise independent.
//
// Caller must hold g.mu.
func (g *Game) seatNeedsTriggerOrderLocked(items []*StackItem, mode TriggerOrderMode) bool {
	if !seatNeedsTriggerOrder(items, mode) {
		return false
	}
	if mode == TriggerOrderAlways || mode == TriggerOrderNever {
		return true
	}
	return !g.triggersIndependentLocked(items)
}

// triggersIndependentLocked reports whether every order of items gives
// the same game: each item is described by its row's footprint, nothing
// on the board reacts to what any of them does, and no two of them
// touch the same thing in a way whose order shows.
//
// Caller must hold g.mu.
func (g *Game) triggersIndependentLocked(items []*StackItem) bool {
	if len(items) < 2 || g.triggerOrderBoardOpaqueLocked() {
		return false
	}
	accs := make([]triggerAccess, len(items))
	for i, t := range items {
		a, ok := g.triggerAccessLocked(t)
		if !ok {
			return false
		}
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
	if g.somethingReactsLocked(emits) || g.writtenObjectsRippleLocked(written) {
		return false
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
			return false
		}
	}
	for i := range accs {
		for j := i + 1; j < len(accs); j++ {
			if g.accessesConflictLocked(&accs[i], &accs[j], eliminable) ||
				g.accessesConflictLocked(&accs[j], &accs[i], eliminable) {
				return false
			}
		}
	}
	return true
}

// accessesConflictLocked reports whether a's writes can change what b
// reads or writes. Called both ways round for each pair.
func (g *Game) accessesConflictLocked(a, b *triggerAccess, eliminable map[uuid.UUID]bool) bool {
	if a.rules {
		return true
	}
	if a.monarch && b.monarch {
		return true
	}
	// A tie on one object asks, whatever the two do to it.
	for id := range a.writes {
		if b.writes[id] || b.reads[id] {
			return true
		}
	}
	for p, dir := range a.life {
		other := b.life[p]
		// Gains commute with gains and losses with losses. A gain
		// and a loss commute too, unless the player could reach 0 in
		// between and lose (CR 704.5a).
		if other != 0 && other != dir && eliminable[p] {
			return true
		}
	}
	for p := range a.cards {
		if b.cards[p] {
			return true
		}
	}
	for p := range eliminable {
		if a.life[p]&lifeDown == 0 && a.draws[p] == 0 {
			continue
		}
		// a can take p out of the game: then whatever b does to p, or
		// to anything p owns or controls, depends on whether b comes
		// first. Another loss of life is the same either way.
		if b.life[p]&lifeUp != 0 || b.cards[p] || b.playerReads[p] {
			return true
		}
		if g.touchesObjectsOfLocked(b, p) {
			return true
		}
	}
	return false
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
	if t == nil || t.Kind != StackItemTriggered || t.Body != CatalogTriggeredBodyKey || t.Params.Ability == nil {
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
	if !reflect.ValueOf(params).IsZero() || len(t.Payload) > 0 || t.XValue != 0 || len(t.Distribution) > 0 {
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
	for _, s := range row.Footprint {
		if !g.addFootprintStepLocked(&a, t, s, targetCards) {
			return triggerAccess{}, false
		}
	}
	for id := range a.writes {
		if !g.objectInertLocked(id) {
			a.rules = true
		}
	}
	return a, true
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
		a.life[you] |= lifeUp
		a.emit(EventChangeLife)
	case FootprintDamageEachOpponent:
		z := g.findCardZoneLocked(t.SourceCardID)
		src, ok := g.LookupCardForEffect(t.SourceCardID)
		if !ok || z == nil || z.Kind != ZoneBattlefield {
			return false
		}
		// Damage that does more than cost life — lifelink (CR
		// 702.15b), infect, wither, toxic — is not on the list.
		tr := damageSourceTraitsOfCard(&src)
		if tr.lifelink || tr.result != (DamageResultSource{}) {
			return false
		}
		a.reads[t.SourceCardID] = true
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == you {
				continue
			}
			a.life[p.ID] |= lifeDown
			a.lifeLoss[p.ID] += s.N
		}
		a.emit(EventDealDamage, EventChangeLife)
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
		a.emit(EventMill, EventZoneMove)
	case FootprintScry:
		a.cards[you] = true
		a.emit(EventScry)
	case FootprintSurveil:
		a.cards[you] = true
		a.emit(EventSurveil, EventZoneMove)
	case FootprintEnergy:
		a.emit(EventPlayerCounterPlaced)
	case FootprintCreateToken:
		if !tokenTemplateInert(s.TokenKey, s.TokenManual) {
			return false
		}
		a.emit(EventTokenCreated, EventETB, EventZoneMove)
	case FootprintMonarch:
		a.monarch = true
		a.emit(EventMonarchChanged)
	default:
		return false
	}
	return true
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
func (g *Game) triggerOrderBoardOpaqueLocked() bool {
	zones := []*Zone{g.Battlefield}
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		if len(p.Statics) > 0 {
			return true
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
				return true
			}
			if d := catalogDef(catalogAbilityKeyOf(c)); d != nil {
				if global, _ := defBearsRules(d); global {
					return true
				}
			}
			for _, t := range triggersOf(c) {
				if t.State != "" {
					return true
				}
			}
		}
	}
	for i := range g.ScopedEffects {
		se := &g.ScopedEffects[i]
		if se.Scope != ScopeNone {
			return true
		}
		for _, m := range se.Mods {
			if !characteristicModKinds[m.Kind] {
				return true
			}
		}
	}
	return false
}

// characteristicModKinds are the continuing-effect mods that only change
// the characteristics or the combat rules of the objects a record locked
// in (layers 1 to 7, and the layer-6 attack and block rules). Whatever
// they do to an object, they do it in every order. Every other mod (a
// prevention or replacement, a "can't", a doubling) changes what an
// event does, and makes the board opaque to the check.
var characteristicModKinds = map[ModKind]bool{
	ModSetController: true, ModAddTypes: true, ModRemoveTypes: true,
	ModAddSubtypes: true, ModAllCreatureTypes: true, ModSetColors: true,
	ModAddKeywords: true, ModRemoveKeywords: true, ModLoseAllAbilities: true,
	ModLoseOwnAbility: true, ModAddRestrictions: true, ModSetBasePower: true,
	ModSetBaseToughness: true, ModModifyPT: true, ModAddAttackRequirement: true,
	ModAddBlockRequirement: true, ModAddBlockCapacity: true, ModBlockAnyNumber: true,
	ModGrantAbilities: true, ModCantAttackUnlessDefenderControls: true,
	ModSetBasicLandTypes: true, ModLoseLandTypes: true, ModCantHaveKeywords: true,
	ModBecomeCopy: true, ModCantBeBlockedExceptBy: true,
	ModLimitBlockersPerDefender: true, ModCantBeBlockedByPlayer: true,
	ModCantBeBlockedByPower: true,
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
func (g *Game) somethingReactsLocked(emits map[EventKind]bool) bool {
	watches := func(kinds []EventKind) bool {
		for _, k := range kinds {
			if emits[k] {
				return true
			}
		}
		return false
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
				if watches(t.Watches) {
					return true
				}
			}
			if CatalogReplacements == nil {
				continue
			}
			for _, r := range CatalogReplacements(catalogAbilityKeyOf(c)) {
				if len(r.Watches) == 0 || watches(r.Watches) {
					return true
				}
			}
		}
	}
	for k := range emits {
		for _, zk := range triggerZones.zonesFor(k) {
			for _, z := range g.triggerZonesOfKindLocked(zk) {
				for i := range z.Cards {
					if triggerZones.declares(CatalogKey(z.Cards[i])) {
						return true
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
				if r.FromGraveyard && (len(r.Watches) == 0 || watches(r.Watches)) {
					return true
				}
			}
		}
	}
	for _, dt := range g.DelayedTriggers {
		if dt != nil && watches(dt.On) {
			return true
		}
	}
	for i := range g.BuiltinReplacements {
		r := &g.BuiltinReplacements[i]
		if !knownBuiltinReplacementLabels[r.Label] && (len(r.Watches) == 0 || watches(r.Watches)) {
			return true
		}
	}
	// CR 702.179d: the active player's speed goes up when an opponent
	// loses life during their turn — an inherent trigger with no card.
	if emits[EventChangeLife] || emits[EventDealDamage] {
		if a := g.Turn.ActiveSeat; a >= 0 && a < len(g.Seats) && g.Seats[a] != nil && g.Seats[a].Speed > 0 {
			return true
		}
	}
	return false
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
	"Resolve": defFieldInert, "AsEnters": defFieldInert, "AsTransformsInto": defFieldInert,
	"StartingLoyalty": defFieldInert, "BattleDefense": defFieldInert,
	"TargetMode": defFieldInert, "Targets": defFieldInert, "Modes": defFieldInert,
	"Purpose": defFieldInert, "ManaAbilities": defFieldInert, "Activated": defFieldInert,
	"EntersWithCountersFromCast": defFieldInert, "PrintedKeywords": defFieldInert,
	"ManaTriggers": defFieldInert, "AdditionalCost": defFieldInert,
	"OptionalCosts": defFieldInert, "AlternativeCosts": defFieldInert,
	"TapCost": defFieldInert, "Delve": defFieldInert, "SpendOnly": defFieldInert,
	"SpendOnlySources": defFieldInert, "SpellsYouCastHaveDelve": defFieldInert,
	"CostModifiers": defFieldInert, "SelfCostModifiers": defFieldInert,
	"ExhaustPermissions": defFieldInert, "BoastLimits": defFieldInert,
	"AttackTaxes": defFieldInert, "BlockRules": defFieldInert, "AttackLimits": defFieldInert,
	"ExertOnAttack": defFieldInert, "CastableZones": defFieldInert,
	"SpecialActions": defFieldInert, "SpecialActionGrants": defFieldInert,
	"UntapStep": defFieldInert, "UntapStepRestrictions": defFieldInert,
	"CounterRemovalLocks": defFieldInert, "UntapCaps": defFieldInert,
	"UntapOptOuts": defFieldInert, "DrawStep": defFieldInert,
	"CastCondition": defFieldInert, "CastConditionLabel": defFieldInert,
	"CastRestrictions": defFieldInert, "LandPlayRestrictions": defFieldInert,
	"ActivationRestrictions": defFieldInert, "ActivationTimings": defFieldInert,
	"CantBeCountered": defFieldInert, "CantBeCounteredIf": defFieldInert,
	"SpellDamageCantBePrevented": defFieldInert, "SpellsCantBeCountered": defFieldInert,
	"HandSize": defFieldInert, "ManaPool": defFieldInert,
	"DamageStaysThroughCleanup": defFieldInert, "AnyColorSpend": defFieldInert,
	"LifeForMana": defFieldInert, "Emblem": defFieldInert, "TokenText": defFieldInert,
	"GrantText": defFieldInert, "XMatters": defFieldInert, "XCeiling": defFieldInert,
	"WantsDistinctColors": defFieldInert, "WantsManaFrom": defFieldInert,
	"AdditionalLandPlays": defFieldInert, "CastPermissions": defFieldInert,
	"GatedCastPermissions": defFieldInert, "CastTimings": defFieldInert,
	"GrantedAlternativeCosts": defFieldInert, "OpeningHand": defFieldInert,
	"LibraryTopVisible": defFieldInert,

	"Triggered": defFieldReactor, "Replacements": defFieldOwnRules,

	"Static": defFieldOwnRules, "TriggerDoublers": defFieldOwnRules,
	"TriggerSuppressors": defFieldOwnRules,

	"HexproofBypasses": defFieldRules, "WardSuppressions": defFieldRules,
	"TargetingRestrictions": defFieldRules, "PlayerKeywords": defFieldRules,
	"PlayerLifeTotalLocked": defFieldRules, "DamageCantBePrevented": defFieldRules,
	"CantGainLife": defFieldRules, "DamageAsThough": defFieldRules,
	"LegendRuleExemptions": defFieldRules, "ZeroLoyaltyExemptions": defFieldRules,
	"OpponentEffectProtections": defFieldRules, "GameEndGates": defFieldRules,
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
		switch class, ok := cardDefFieldClass[f.Name]; {
		case !ok || class == defFieldRules:
			global = true
		case class == defFieldOwnRules:
			own = true
		}
	}
	return global, own
}
