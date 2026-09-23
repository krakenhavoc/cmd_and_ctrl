package game

import "github.com/google/uuid"

// TriggerDoubler is a catalog-declared CR 603.2d count modifier for a
// triggered ability. Each matching doubler adds one instance.
type TriggerDoubler struct {
	Label   string
	Applies func(g *Game, q TriggerDoublingQuery) bool
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
type doublerCandidate struct {
	card     Card
	lki      Characteristic
	doublers []TriggerDoubler
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
	doublers   []doublerCandidate
	scanned    bool
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
		if entering && g.hasTriggerDoublerLocked() {
			// Entry invalidates the layer cache before the harvester runs.
			// Capture the entering permanent with continuous effects applied:
			// Mycosynth Lattice makes even a Forest enter as an artifact.
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

// Only materialize entry characteristics when a catalog doubler can use them.
// Read the printed slot even for a currently silenced permanent: entry can
// change continuous effects, so the stale cache cannot decide ability removal.
func (g *Game) hasTriggerDoublerLocked() bool {
	if g.Battlefield == nil || CatalogTriggerDoublers == nil {
		return false
	}
	for _, c := range g.Battlefield.Cards {
		if key := CatalogKey(c); key != "" && len(CatalogTriggerDoublers(key)) > 0 {
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
		g.scanTriggerDoublersLocked(p)
	}
	if len(p.doublers) == 0 {
		return nil
	}
	decl := ability
	var out []doublerRef
	for _, candidate := range p.doublers {
		if candidate.lki.AbilitiesRemoved {
			continue
		}
		for _, doubler := range candidate.doublers {
			if doubler.Applies == nil {
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

func (g *Game) scanTriggerDoublersLocked(p *harvestPass) {
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
		doublers := CatalogTriggerDoublers(key)
		if len(doublers) == 0 {
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
		p.doublers = append(p.doublers, doublerCandidate{card: c, lki: lki, doublers: doublers})
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
