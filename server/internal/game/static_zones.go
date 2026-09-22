package game

// static_zones.go — #1221: WHERE a static ability functions from.
//
// CR 113.6 again, and this is its third reader. The activated half is
// ability_zone.go (#660, ADR 0062 Decision 1), the triggered half is
// trigger_zones.go (#925), and until this file the LAYER pass had one
// answer for everything: `activeStaticAbilitiesLocked` walked
// `g.Battlefield` and nothing else, so a static printed on a card in a
// graveyard was never gathered, never sorted into a bucket and never
// applied.
//
// CR 113.6c is the clause the cards quote:
//
//	"As long as Anger is in your graveyard and you control a
//	 Mountain, creatures you control have haste."
//
// The Judgment incarnations (Anger, Wonder, Brawn, Valor) are the
// family; Yixlid Jailer's "cards in graveyards lose all abilities" is
// a battlefield static about graveyards and is NOT this — the
// difference is which zone the SOURCE is in.
//
// # The index, and why the battlefield does not get slower
//
// A blanket "walk every graveyard on every layer recompute" would be
// a real cost on a four-player table with sixty cards in the yards,
// paid whether or not anything in any deck functions from there. So
// the catalog is indexed ONCE, at effects.Register, exactly as #925
// indexes trigger zones:
//
//   - `zones()` answers "which non-battlefield zones does any
//     registered card declare a static from". The usual answer is
//     none and the whole walk is then one slice read.
//   - `declares(oracleKey)` answers "does this card declare any
//     non-battlefield static at all", so a walked zone skips the
//     catalog lookup for every card that does not.
//
// The battlefield walk gains exactly one comparison per ability
// (`StaticFunctionsFromZone`), which is what keeps a declared zone
// from ALSO applying on the battlefield.
//
// # Contract of the index
//
// Written at boot from effects.Register, read-only afterwards — the
// contract every Catalog* hook carries, and the reason a test may
// swap it wholesale the way those tests swap CatalogStaticAbilities.

// defaultStaticZones is what a static that declares nothing gets: the
// battlefield, and nowhere else. That is every static the catalog
// held before #1221.
var defaultStaticZones = []ZoneKind{ZoneBattlefield}

// supportedStaticZones is the set of non-battlefield zones the layer
// gather walks. The GRAVEYARD alone, because it is the only one a
// printed card asks for (the four incarnations, CR 113.6c), and a
// zone nothing walks would be a declaration the engine silently
// ignored. effects.Register refuses the rest at boot.
//
// The hand is the obvious next one — "as long as this card is in your
// hand" is a real printed clause — and it is one entry here plus one
// line in zonesOfKindLocked's existing switch on the day a card needs
// it.
var supportedStaticZones = []ZoneKind{ZoneGraveyard}

// StaticZones is the zones a static ability functions from. Never
// empty.
func StaticZones(s StaticAbility) []ZoneKind {
	if len(s.Zones) == 0 {
		return defaultStaticZones
	}
	return s.Zones
}

// StaticFunctionsFromZone reports whether `s` applies while its source
// sits in `zone` (CR 113.6).
//
// Note what it does NOT do, for the reason AbilityFunctionsFromZone
// and TriggerWatchesFromZone do not: a static that declares the
// graveyard does not thereby also apply from the battlefield. Anger's
// anthem is printed as a graveyard clause and its own haste is a
// separate printed keyword; a source list that silently added the
// battlefield would give every creature haste while Anger is in play,
// which the card does not say.
func StaticFunctionsFromZone(s StaticAbility, zone ZoneKind) bool {
	for _, z := range StaticZones(s) {
		if z == zone {
			return true
		}
	}
	return false
}

// StaticZoneUnsupported names the reason `zone` cannot be declared on
// a static ability, or "" when it can. Used by effects.Register to
// refuse a declaration at boot rather than leave a card with a static
// nothing ever gathers.
func StaticZoneUnsupported(zone ZoneKind) string {
	// Read off supportedStaticZones rather than off a second switch,
	// so adding a zone to the gather's list is the ONE edit that
	// opens it — a hand-maintained twin here is a zone that either
	// registers and is never walked, or is walked and never
	// registers.
	for _, z := range supportedStaticZones {
		if z == zone {
			return ""
		}
	}
	if zone == ZoneBattlefield {
		return "the battlefield is what an empty Zones means — declare nothing"
	}
	return "the layer gather walks only " + joinZoneKinds(supportedStaticZones) +
		" (#1221); add the zone to supportedStaticZones with the card that needs it"
}

// joinZoneKinds renders a zone list for an error message.
func joinZoneKinds(zones []ZoneKind) string {
	out := ""
	for i, z := range zones {
		switch {
		case i == 0:
		case i == len(zones)-1:
			out += " and "
		default:
			out += ", "
		}
		out += string(z)
	}
	return out
}

// staticZoneIndex is the catalog-wide index of non-battlefield static
// declarations. See the file header for the contract.
type staticZoneIndex struct {
	// zoneList is every non-battlefield zone some registered card
	// declares a static from, deduplicated, in supportedStaticZones
	// order.
	zoneList []ZoneKind
	// oracles is every catalog key that declares at least one
	// non-battlefield static.
	oracles map[string]bool
}

func newStaticZoneIndex() *staticZoneIndex {
	return &staticZoneIndex{oracles: map[string]bool{}}
}

// staticZones is the process-wide index. Package-level, like the
// Catalog* hooks, and written only from IndexStaticZones.
var staticZones = newStaticZoneIndex()

// IndexStaticZones records the non-battlefield zones `statics`
// function from, under `oracleKey`. Called once per card from
// effects.Register; a card with no such static costs one loop and
// stores nothing.
//
// Idempotent for a given key and static list — re-registering the
// same card re-adds the same entries.
func IndexStaticZones(oracleKey string, statics []StaticAbility) {
	staticZones.index(oracleKey, statics)
}

// index is IndexStaticZones against one index rather than the
// process-wide one, so a test can exercise the contract without
// touching boot-time state it would then have to put back.
func (ix *staticZoneIndex) index(oracleKey string, statics []StaticAbility) {
	if oracleKey == "" {
		return
	}
	for _, s := range statics {
		if len(s.Zones) == 0 {
			continue
		}
		ix.oracles[oracleKey] = true
		for _, zone := range s.Zones {
			if zone == ZoneBattlefield {
				continue
			}
			ix.add(zone)
		}
	}
}

// add records "the layer gather must walk this zone".
func (ix *staticZoneIndex) add(zone ZoneKind) {
	for _, z := range ix.zoneList {
		if z == zone {
			return
		}
	}
	ix.zoneList = append(ix.zoneList, zone)
}

// zones is the per-recompute question: which non-battlefield zones
// have to be gathered at all. Nil for almost every game.
func (ix *staticZoneIndex) zones() []ZoneKind { return ix.zoneList }

// declares reports whether a catalog key has any non-battlefield
// static, so a walked zone can skip the catalog lookup for the cards
// that do not.
func (ix *staticZoneIndex) declares(oracleKey string) bool {
	return oracleKey != "" && ix.oracles[oracleKey]
}

// declaredZoneStaticsLocked is the ONE narrow gather #1221 adds: for
// each non-battlefield zone the index says something functions from,
// walk that zone and bind the statics that declared it.
//
// Three details are the whole of the rule:
//
//  1. **The source is the card AS IT SITS IN THAT ZONE**, copied by
//     value with `Controller = Owner`. CR 108.4: a card outside the
//     battlefield and the stack has no controller, and its OWNER is
//     the "you" of its abilities — so Anger's "creatures YOU control"
//     means its owner's, and an Anger milled out of an opponent's
//     library gives that opponent nothing. The same sentence
//     harvestFromDeclaredZone applies to triggers and
//     ActivateCatalogAbility applies to activations.
//
//  2. **`live` is false.** It marks a source that is a permanent on
//     the battlefield right now, and is read for one thing:
//     CR 613.1f ability-removal silencing. Nothing on the board can
//     silence a card in a graveyard — Darksteel Mutation applies to
//     a permanent — so a graveyard static is not a candidate, and
//     saying so here is what keeps applyBucketLocked from asking.
//
//  3. **The timestamp is the source's last battlefield entry**, which
//     is when it died for the card this exists for and zero for one
//     milled straight out of a library. CR 613.7 wants the time the
//     object entered the zone it is in, and no field records that. It
//     is unobservable for this whole family — every incarnation is a
//     layer-6 keyword GRANT, and grants commute — so the choice is
//     between an approximation and a new field on Card that the
//     snapshot, the clone and the drift guard would all have to carry
//     for a number nothing can read. Stated rather than hidden: a
//     graveyard static that SET a characteristic would need the real
//     thing, and there is none.
//
// Caller must hold g.mu.
func (g *Game) declaredZoneStaticsLocked() []ContinuousEffect {
	if CatalogStaticAbilities == nil {
		return nil
	}
	var out []ContinuousEffect
	for _, kind := range staticZones.zones() {
		for _, z := range g.zonesOfKindLocked(kind) {
			if z == nil {
				continue
			}
			for i := range z.Cards {
				// The oracle key is the pre-filter's question — "does
				// any card with this key declare a zone at all" — and
				// it is answered first because it is one map read and
				// it skips the whole of every ordinary card.
				if !staticZones.declares(CatalogKey(z.Cards[i])) {
					continue
				}
				src := z.Cards[i]
				src.Controller = src.Owner
				// StaticAbilitiesForCard, not the raw hook: this is
				// the second place an object becomes statics and it
				// goes through the same accessor (ADR 0071), so a
				// designation gate is evaluated here too. Off the
				// battlefield CR 400.7 has already cleared every
				// designation, which is the right answer and one a
				// parallel read here would have got wrong.
				for _, ab := range StaticAbilitiesForCard(src) {
					if !StaticFunctionsFromZone(ab, kind) {
						continue
					}
					bound := src
					out = append(out, staticContinuousEffect{
						ability:   ab,
						source:    &bound,
						timestamp: bound.EnteredBattlefieldAt,
						live:      false,
					})
				}
			}
		}
	}
	return out
}
