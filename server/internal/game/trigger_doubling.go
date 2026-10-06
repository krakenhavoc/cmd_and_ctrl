package game

import "github.com/google/uuid"

// TriggerDoubler is a catalog-declared CR 603.2d count modifier for a
// triggered ability. Each matching doubler adds one instance.
type TriggerDoubler struct {
	Label   string
	Applies func(g *Game, q TriggerDoublingQuery) bool

	// ActiveWhen is the ADR 0071 designation gate: this doubler is
	// active only while its source permanent has the named
	// designation. The zero value (DesignationAlways) means "always
	// active" — every doubler declared before #1647.
	//
	// A TriggerDoubler is not an ability-list entry (it is consulted
	// directly by triggerDoublersLocked, never through
	// TriggersForCard / StaticAbilitiesForCard), so it cannot reach
	// the ordinary designation gate the way a TriggeredAbility or
	// StaticAbility does. This field is that gate's own doorway onto
	// the doubler: Windcrag Siege's Mardu mode sets
	// ActiveWhen: ChosenIs("Mardu"), so the doubler exists only while
	// "Mardu" is the permanent's chosen option (Card.ChosenOption),
	// and — because DesignationChosenOption never matches an empty
	// answer — not at all before the controller answers the as-enters
	// prompt.
	//
	// Checked against the DOUBLER's own card (TriggerDoublingQuery.Doubler),
	// the permanent this ability is printed on, never against Source
	// (the permanent whose trigger is being doubled).
	ActiveWhen Designation
}

type TriggerDoublingQuery struct {
	Event      Event
	Doubler    Card
	DoublerLKI Characteristic
	Source     Card
	SourceLKI  Characteristic
	FromSpell  bool
	Ability    *TriggeredAbility
	Subject    uuid.UUID
	SubjectLKI Characteristic
	HasSubject bool
}

type doublerRef struct {
	id   uuid.UUID
	name string
}

// modifierCandidate is one permanent that declares a CR 603.2d doubler,
// a trigger suppressor (trigger_suppression.go), or both, as one
// event's harvest sees it.
type modifierCandidate struct {
	card        Card
	lki         Characteristic
	doublers    []TriggerDoubler
	suppressors []TriggerSuppressor
	// live reports that the candidate is on the battlefield as the
	// event is harvested. A doubler never reads it (ADR 0018 Decision
	// 3 counts a doubler that leaves in the same event). A suppressor
	// does, away from a leaves-the-battlefield event: there the game
	// reads the board AFTER the event (CR 603.10), so a suppressor that
	// has already left is not there to stop anything.
	live bool
}

// triggerIdentityLKI is the non-characteristic identity a trigger needs
// after CR 400.7 has restored the destination card's printed self.
type triggerIdentityLKI struct {
	OracleID string `json:"oracleID,omitempty"`
	// TokenKey is OracleID's sibling for a TOKEN, whose catalog
	// identity is its template's synthetic key rather than an oracle
	// ID (#521, token_key.go). Recorded for exactly the reason the
	// oracle ID is: the identity a trigger is read off must be the
	// one the permanent had while it was still on the battlefield
	// (CR 603.10), and for a token that identity is this field.
	// Without it, restoring the identity would BLANK a dying token's
	// key and its "when this token dies" would never be found.
	// ADR 0083 decision 4.
	TokenKey   string    `json:"tokenKey,omitempty"`
	ActiveFace int       `json:"activeFace,omitempty"`
	AttachedTo TargetRef `json:"attachedTo,omitempty"`
}

// harvestPass is the event-local, immutable view shared by every matching
// trigger. Candidate scanning is deferred until an ability actually matches.
type harvestPass struct {
	ev         Event
	subject    uuid.UUID
	subjectLKI Characteristic
	hasSubject bool
	modifiers  []modifierCandidate
	scanned    bool
	// settled is set by settleBatchEndTriggersLocked: the batch is over,
	// so an AtBatchEnd ability dispatches instead of staging (#2183).
	settled bool
}

func (g *Game) newHarvestPassLocked(ev Event) harvestPass {
	p := harvestPass{ev: ev}
	if isBattlefieldExitEvent(ev) && ev.CardID != uuid.Nil {
		if c, ok := g.simultaneousExitCardLocked(ev.CardID); ok {
			p.subject, p.subjectLKI, p.hasSubject = ev.CardID, c.Effective(), true
			return p
		}
		if lki, ok := g.lastKnownBattlefield[ev.CardID]; ok {
			p.subject, p.subjectLKI, p.hasSubject = ev.CardID, lki, true
		}
		return p
	}
	// Combat damage names its subject through Source, not CardID (which
	// other damage consumers may use for a different purpose).
	if ev.Kind == EventDealDamage && ev.Combat && ev.Source != uuid.Nil {
		if c := g.findCardByIDLocked(ev.Source); c != nil {
			p.subject, p.subjectLKI, p.hasSubject = c.InstanceID, c.Effective(), true
		}
		return p
	}
	entering := ev.Kind == EventETB || ev.Kind == EventTokenCreated || (ev.Kind == EventZoneMove && ev.NewZone == ZoneBattlefield)
	needsCardSubject := entering || ev.Kind == EventAttack || ev.Kind == EventCast
	if needsCardSubject && ev.CardID != uuid.Nil {
		if entering && g.hasTriggerModifierLocked() {
			// Entry invalidates the layer cache before the harvester runs.
			// Capture the entering permanent with continuous effects applied:
			// Mycosynth Lattice makes even a Forest enter as an artifact,
			// and Torpor Orb has to see a land that enters animated as the
			// creature it is (CR 603.6a).
			// Exit events above keep their pre-move characteristics instead.
			g.RecomputeLayersIfStaleLocked()
		}
		if c := g.findCardByIDLocked(ev.CardID); c != nil {
			p.subject, p.subjectLKI, p.hasSubject = c.InstanceID, c.Effective(), true
			return p
		}
	}
	return p
}

// Only materialize entry characteristics when a catalog doubler or
// suppressor can use them. Read the printed slot even for a currently
// silenced permanent: entry can change continuous effects, so the stale
// cache cannot decide ability removal.
func (g *Game) hasTriggerModifierLocked() bool {
	if g.Battlefield == nil || (CatalogTriggerDoublers == nil && CatalogTriggerSuppressors == nil) {
		return false
	}
	for _, c := range g.Battlefield.Cards {
		key := CatalogKey(c)
		if key == "" {
			continue
		}
		if CatalogTriggerDoublers != nil && len(CatalogTriggerDoublers(key)) > 0 {
			return true
		}
		if CatalogTriggerSuppressors != nil && len(CatalogTriggerSuppressors(key)) > 0 {
			return true
		}
	}
	return false
}

func isBattlefieldExitEvent(ev Event) bool {
	return ev.Kind == EventLTB || (ev.Kind == EventZoneMove && ev.OldZone == ZoneBattlefield)
}

// simultaneousExitCardLocked returns the pre-move copy for one member of
// the current simultaneous exit. It is deliberately event-local: callers
// must not retain the returned card after the batch closes.
func (g *Game) simultaneousExitCardLocked(id uuid.UUID) (Card, bool) {
	for _, c := range g.simultaneousExit {
		if c.InstanceID == id {
			return c, true
		}
	}
	return Card{}, false
}

func (g *Game) triggerDoublersLocked(p *harvestPass, source Card, lki Characteristic, ability TriggeredAbility, fromSpell bool) []doublerRef {
	if CatalogTriggerDoublers == nil {
		return nil
	}
	if !p.scanned {
		g.scanTriggerModifiersLocked(p)
	}
	if len(p.modifiers) == 0 {
		return nil
	}
	decl := ability
	var out []doublerRef
	for _, candidate := range p.modifiers {
		if candidate.lki.AbilitiesRemoved || len(candidate.doublers) == 0 {
			continue
		}
		for _, doubler := range candidate.doublers {
			if doubler.Applies == nil {
				continue
			}
			if doubler.ActiveWhen.IsGate() && !doubler.ActiveWhen.Active(candidate.card) {
				continue
			}
			q := TriggerDoublingQuery{Event: p.ev, Doubler: candidate.card, DoublerLKI: candidate.lki, Source: source, SourceLKI: lki, FromSpell: fromSpell, Ability: &decl, Subject: p.subject, SubjectLKI: p.subjectLKI, HasSubject: p.hasSubject}
			if doubler.Applies(g, q) {
				name := doubler.Label
				if name == "" {
					name = candidate.card.Name
				}
				out = append(out, doublerRef{id: candidate.card.InstanceID, name: name})
			}
		}
	}
	return out
}

// scanTriggerModifiersLocked collects, once per event, every permanent
// that declares a doubler or a suppressor. One walk serves both: the
// suppressors are asked about every match before the doublers are
// (harvestMatchLocked), so a second walk would double the cost of every
// event that matches a trigger at all.
func (g *Game) scanTriggerModifiersLocked(p *harvestPass) {
	p.scanned = true
	// A normal harvest only walks the battlefield, whose identities are
	// unique. The dedupe table is needed only while a simultaneous-exit
	// snapshot overlaps that zone; keep the steady-state hot path allocation
	// free.
	var seen map[uuid.UUID]bool
	add := func(c Card, known *Characteristic, authoritative bool) {
		if c.InstanceID == uuid.Nil || (seen != nil && seen[c.InstanceID]) {
			return
		}
		// A simultaneous-exit copy is the authoritative view for this
		// identity. Mark it even if it has lost its abilities: falling back
		// to the destination copy later would incorrectly rediscover the
		// catalog slot after the LKI has been consumed.
		if authoritative {
			if seen == nil {
				seen = make(map[uuid.UUID]bool, len(g.simultaneousExit))
			}
			seen[c.InstanceID] = true
		}
		if known != nil && known.AbilitiesRemoved {
			return
		}
		key := CatalogAbilityKey(c)
		if key == "" {
			return
		}
		var doublers []TriggerDoubler
		if CatalogTriggerDoublers != nil {
			doublers = CatalogTriggerDoublers(key)
		}
		var suppressors []TriggerSuppressor
		if CatalogTriggerSuppressors != nil {
			suppressors = CatalogTriggerSuppressors(key)
		}
		if len(doublers) == 0 && len(suppressors) == 0 {
			return
		}
		// Effective characteristics are comparatively expensive: the normal
		// path sees far more permanents than actual trigger doublers. Query
		// the catalog slot first and only materialize LKI for a candidate.
		lki := c.Effective()
		if known != nil {
			lki = *known
		}
		if lki.AbilitiesRemoved {
			return
		}
		if seen != nil {
			seen[c.InstanceID] = true
		}
		p.modifiers = append(p.modifiers, modifierCandidate{
			card: c, lki: lki, doublers: doublers, suppressors: suppressors,
			live: findCardOnBattlefield(g, c.InstanceID) >= 0,
		})
	}
	// A simultaneous-exit copy wins over a still-live battlefield card with
	// the same ID: another member of the wipe may already have removed a
	// continuous effect, while the batch copy is the characteristics that
	// existed for the whole simultaneous exit.
	for _, c := range g.simultaneousExit {
		add(c, nil, true)
	}
	if g.Battlefield != nil {
		for _, c := range g.Battlefield.Cards {
			add(c, nil, false)
		}
	}
	if isBattlefieldExitEvent(p.ev) && p.ev.CardID != uuid.Nil {
		if c := g.findCardByIDLocked(p.ev.CardID); c != nil {
			leaving := g.withLastKnownTriggerIdentityLocked(*c)
			lki := c.Effective()
			if p.hasSubject && p.subject == p.ev.CardID {
				lki = p.subjectLKI
			} else if known, ok := g.lastKnownBattlefield[p.ev.CardID]; ok {
				lki = known
			}
			// An ability-removing effect applies to the leaving permanent too.
			if !lki.AbilitiesRemoved {
				add(leaving, &lki, false)
			}
		}
	}
}

func (g *Game) withLastKnownTriggerIdentityLocked(c Card) Card {
	identity, ok := g.lastKnownTriggerIdentity[c.InstanceID]
	if !ok {
		return c
	}
	c.OracleID = identity.OracleID
	c.TokenKey = identity.TokenKey
	c.ActiveFace = identity.ActiveFace
	c.AttachedTo = identity.AttachedTo
	if lki, ok := g.lastKnownBattlefield[c.InstanceID]; ok {
		c.Name = lki.Name
		c.Controller = lki.Controller
		c.effective = &lki
	}
	return c
}
