package game

import (
	"sort"
	"sync"
)

// zone_grants.go — abilities granted to CARDS that are not on the
// battlefield (ADR 0093, amendment 2026-10-13; #3233).
//
//	CR 113.10  "Some effects give an object an ability."
//	CR 613.1f  "Layer 6: Ability-adding effects … are applied."
//
// Satoru Umezawa's "Each creature card in your hand has ninjutsu
// {2}{U}{B}." is the card. Until this file a grant (StaticAbility.
// GrantAbilities) reached permanents alone: the layer pass computes a
// characteristic for a battlefield card and for nothing else, so a card
// in a hand had no place to hold a granted row.
//
// # The shape
//
// A static declares StaticAbility.GrantZone beside GrantAbilities. After
// the battlefield pass has settled control (layer 2) and each source's
// own ability removal (layer 6), this step walks the named zone of every
// seat and, for each card the static's AppliesTo accepts, appends one
// GrantedAbility (the bundle key and the grantor, the same plain data a
// permanent's grant is) to Card.zoneGranted, in the source's timestamp
// order (CR 613.7). Nothing else is written anywhere.
//
// # How a reader finds it
//
// Through the ONE ability-row accessor, activatedAbilityRows, which
// reads zoneGranted where a permanent's rows read its layered grants
// (abilityGrantsOf). So the activation path, the wire's zone_abilities
// row, the legal-move enumerator and the lobby lookup all see a granted
// ninjutsu with no change of their own, and the row's ref
// ("grant:<bundle>:<i>:<n>") is what an announcement and a restored
// stack item name, exactly as for a permanent. The bundle's own
// declaration says where it functions (ActivatedAbility.Zones), and
// AbilityFunctionsFromZone holds the activation to it as it does a
// printed cycling row.
//
// # When it ends
//
// The grant is re-derived on every pass, so it ends the pass after its
// source leaves the battlefield or loses its abilities (CR 113.6,
// 613.6), and a card that changes zone drops it at once (MoveCard).
// The layer version is bumped when a card ENTERS a zone a live zone
// grant covers (layerVersionBump), because that move alone would not
// otherwise invalidate the pass: a creature card drawn while Satoru is
// out has the ability from the moment it is in the hand.
//
// A source silenced by a CR 613.1f removal contributes nothing, judged
// on the battlefield pass's final answer for the pass, as the spell
// statics are (spell_keywords.go). The one case that is not exact is a
// grantor whose own static is newer than the effect silencing it
// (CR 613.7); the catalog has no pair of cards that makes the
// difference observable, and judging it exactly would mean running this
// step inside the bucket loop.

// supportedGrantZones are the zones whose cards a static may grant to.
// The library is missing on purpose: it is hidden, and no printed
// effect gives its cards abilities. The command zone is missing because
// its only occupants are commanders, whose own text names it.
var supportedGrantZones = []ZoneKind{ZoneHand, ZoneGraveyard, ZoneExile}

// GrantZoneUnsupported names the reason `zone` cannot be a
// StaticAbility.GrantZone, or "" when it can. Used by effects.Register
// to refuse a declaration the pass would silently ignore.
func GrantZoneUnsupported(zone ZoneKind) string {
	for _, z := range supportedGrantZones {
		if z == zone {
			return ""
		}
	}
	if zone == ZoneBattlefield {
		return "the battlefield is what an empty GrantZone means — declare nothing"
	}
	return "a zone grant walks only " + joinZoneKinds(supportedGrantZones) +
		" (ADR 0093 amendment 2026-10-13); add the zone to supportedGrantZones with the card that needs it"
}

// zoneGrantIndex is the catalog-wide set of keys that declare a zone
// grant, so the pass pays one map read per permanent rather than a
// catalog walk. Written at boot from effects.Register, read-only
// afterwards — the contract every Catalog* hook carries.
var zoneGrantIndex = struct {
	sync.RWMutex
	keys map[string]bool
}{keys: map[string]bool{}}

// IndexZoneGrants records that `key` declares a zone grant. Called once
// per card from effects.Register; a card with none stores nothing.
func IndexZoneGrants(key string, statics []StaticAbility) {
	if key == "" {
		return
	}
	for _, s := range statics {
		if s.GrantZone != "" {
			zoneGrantIndex.Lock()
			zoneGrantIndex.keys[key] = true
			zoneGrantIndex.Unlock()
			return
		}
	}
}

func declaresZoneGrant(key string) bool {
	zoneGrantIndex.RLock()
	ok := zoneGrantIndex.keys[key]
	zoneGrantIndex.RUnlock()
	return ok
}

// zoneGrantSource is one battlefield static that grants to cards, bound
// to the permanent it comes from.
type zoneGrantSource struct {
	ability StaticAbility
	source  *Card
	ts      int64
}

// zoneGrantSourcesLocked gathers the zone-granting statics on the
// battlefield, in timestamp order. Nil, and no allocation, on a board
// with none, which is nearly every board.
//
// Caller must hold g.mu (write); the source pointers are into the
// battlefield slice and are good only for this pass.
func (g *Game) zoneGrantSourcesLocked() []zoneGrantSource {
	if g.Battlefield == nil || CatalogStaticAbilities == nil {
		return nil
	}
	var out []zoneGrantSource
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if !declaresZoneGrant(catalogKeyOf(src)) || src.HasLostAllAbilities() {
			continue
		}
		for _, ab := range staticAbilitiesOf(src) {
			if ab.GrantZone != "" && ab.AppliesTo != nil && len(ab.GrantAbilities) > 0 {
				out = append(out, zoneGrantSource{ability: ab, source: src, ts: src.layerTimestamp()})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ts < out[j].ts })
	return out
}

// zoneGrantPassLocked is the zone grant step: it re-derives
// Card.zoneGranted for every card in a hand, a graveyard or exile. It
// runs after the battlefield pass and layer 2 (recomputeLayersLocked),
// because "your hand" reads the grantor's controller.
//
// It runs unconditionally once any card in the catalog declares a zone
// grant, because ending a grant is a write too: the walk clears what the
// last pass wrote. That is a nil test per card in the three zones.
//
// Caller must hold g.mu (write).
func (g *Game) zoneGrantPassLocked() {
	zoneGrantIndex.RLock()
	declared := len(zoneGrantIndex.keys) > 0
	zoneGrantIndex.RUnlock()
	if !declared {
		return
	}
	sources := g.zoneGrantSourcesLocked()
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		g.grantInPileLocked(p.Hand, ZoneHand, sources)
		g.grantInPileLocked(p.Graveyard, ZoneGraveyard, sources)
	}
	g.grantInPileLocked(g.Exile, ZoneExile, sources)
}

// grantInPileLocked is zoneGrantPassLocked's per-pile body.
func (g *Game) grantInPileLocked(z *Zone, kind ZoneKind, sources []zoneGrantSource) {
	if z == nil {
		return
	}
	for i := range z.Cards {
		c := &z.Cards[i]
		var out []GrantedAbility
		for _, s := range sources {
			if s.ability.GrantZone != kind || !s.ability.AppliesTo(c, g, s.source) {
				continue
			}
			for _, key := range s.ability.GrantAbilities {
				if gk := GrantKey(key); gk != "" {
					out = append(out, GrantedAbility{Key: gk, Source: s.source.InstanceID, Activator: s.source.Controller})
				}
			}
		}
		c.zoneGranted = out
	}
}

// zoneGrantStaticIsLiveLocked reports whether any permanent on the
// battlefield declares a zone grant into `zone`. The gate on the bump in
// layerVersionBump for a card entering that zone.
//
// Caller must hold g.mu.
func zoneGrantStaticIsLiveLocked(g *Game, zone ZoneKind) bool {
	return staticOnBattlefieldLocked(g, func(ab StaticAbility) bool { return ab.GrantZone == zone })
}

// abilityGrantsOf is the grants an object's ability rows are built
// from: a permanent's layer-6 grants, or, off the battlefield, the zone
// grants the last pass stamped on the card. A permanent has no zone
// grants and a card has no layered ones, so the two never mix.
func abilityGrantsOf(c *Card) []GrantedAbility {
	if c.effective != nil {
		return c.effective.GrantedAbilities
	}
	return c.zoneGranted
}
