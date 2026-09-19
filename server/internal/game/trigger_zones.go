package game

// trigger_zones.go — #925: WHERE a triggered ability watches from.
//
// CR 113.6: an ability of a card that is not on the battlefield
// functions only if it says it does. "When you cycle this card, it
// deals 1 damage to each opponent" says so, and the card is in a
// graveyard by the time the ability triggers (CR 702.29c — that is
// the rule, not an accident). Suspend says so twice, from exile
// (CR 702.62b/c). Bloodghast's landfall says so from a graveyard.
//
// Until this file the harvester had one answer: `triggerHarvester`
// walked `g.Battlefield` and two narrow special cases beside it —
// `harvestCastFromStack` for "when you cast this spell" and
// `harvestLTB` for the card whose battlefield exit is being
// reported. Neither is a zone dimension: the first is one card the
// event names, the second is one card on its way out.
//
// So `TriggeredAbility.Zones` is the sibling of
// `ActivatedAbilityShape.Zones` (ADR 0062 Decision 1) and of
// `CastableZones`: per-declaration, nil meaning the overwhelmingly
// common answer, and read by ONE narrow harvest rather than by a
// blanket scan of every zone.
//
// # The index, and why the battlefield does not get slower
//
// A blanket "walk every graveyard on every event" would be a real
// cost on a four-player game with sixty cards in the yards, paid by
// every table whether or not anything in the deck watches from
// there. Instead the catalog is indexed ONCE, at
// `effects.Register`:
//
//   - `zonesFor(kind)` answers "which non-battlefield zones does any
//     registered card watch THIS event kind from". The usual answer
//     is none, and the whole walk is then one map lookup.
//   - `declares(oracleKey)` answers "does this card declare any
//     non-battlefield trigger at all", so a walked zone skips the
//     catalog lookup for every card that does not.
//
// The battlefield walk gains exactly one comparison per ability
// (`watchesFromZone`), which is what keeps a declared zone from ALSO
// firing from the battlefield — the same comparison `harvestLTB` and
// the simultaneous-exit pass make, so there is one rule about where
// an ability lives and three readers of it rather than three rules.
//
// # Contract of the index
//
// Written at boot from `effects.Register`, read-only afterwards —
// exactly the contract `CatalogTriggers` and the rest of the
// `Catalog*` hooks carry, and the reason a test may swap it wholesale
// the way those tests swap `CatalogTriggers`.

// defaultTriggerZones is what a triggered ability that declares
// nothing gets: the battlefield, and nowhere else. That is every
// ability the catalog held before #925.
var defaultTriggerZones = []ZoneKind{ZoneBattlefield}

// supportedTriggerZones is the set of non-battlefield zones the
// harvest walks. Graveyard and exile are the three keyword families
// #925 was opened for (cycling-this-card, suspend, graveyard
// recursion); the hand and the library have no card asking yet, and
// a zone nothing walks would be a declaration the engine silently
// ignored. `effects.Register` refuses the rest at boot.
var supportedTriggerZones = []ZoneKind{ZoneGraveyard, ZoneExile}

// TriggerZones is the zones a triggered ability watches from. Never
// empty.
func TriggerZones(t TriggeredAbility) []ZoneKind {
	if len(t.Zones) == 0 {
		return defaultTriggerZones
	}
	return t.Zones
}

// TriggerWatchesFromZone reports whether `t` watches events while its
// source sits in `zone` (CR 113.6).
//
// Note what it does NOT do, for the same reason
// `AbilityFunctionsFromZone` does not: an ability that declares the
// graveyard does not thereby also watch from the battlefield. "When
// you cycle this card" cannot fire off a permanent — there is no card
// in hand to discard — and a declared zone list IS the list.
func TriggerWatchesFromZone(t TriggeredAbility, zone ZoneKind) bool {
	for _, z := range TriggerZones(t) {
		if z == zone {
			return true
		}
	}
	return false
}

// TriggerZoneUnsupported names the reason `zone` cannot be declared
// on a triggered ability, or "" when it can. Used by
// `effects.Register` to refuse a declaration at boot rather than
// leave a card with a trigger nothing ever walks.
func TriggerZoneUnsupported(zone ZoneKind) string {
	switch zone {
	case ZoneGraveyard, ZoneExile:
		return ""
	case ZoneBattlefield:
		return "the battlefield is what an empty Zones means — declare nothing"
	case ZoneStack:
		return "a trigger on a spell on the stack is TriggeredAbility.FromStack (cascade, CR 702.85a)"
	default:
		return "only the graveyard and exile are walked (#925); add the zone to supportedTriggerZones with the card that needs it"
	}
}

// triggerZoneIndex is the catalog-wide index of non-battlefield
// trigger declarations. See the file header for the contract.
type triggerZoneIndex struct {
	// zones[kind] is every non-battlefield zone some registered card
	// watches `kind` from, deduplicated, in supportedTriggerZones
	// order.
	zones map[EventKind][]ZoneKind
	// oracles is every catalog key that declares at least one
	// non-battlefield trigger.
	oracles map[string]bool
}

func newTriggerZoneIndex() *triggerZoneIndex {
	return &triggerZoneIndex{
		zones:   map[EventKind][]ZoneKind{},
		oracles: map[string]bool{},
	}
}

// triggerZones is the process-wide index. Package-level, like the
// Catalog* hooks, and written only from IndexTriggerZones.
var triggerZones = newTriggerZoneIndex()

// IndexTriggerZones records the non-battlefield zones `triggers`
// watch from, under `oracleKey`. Called once per card from
// `effects.Register`; a card with no such trigger costs one loop and
// stores nothing.
//
// Idempotent for a given key and trigger list — re-registering the
// same card re-adds the same entries.
func IndexTriggerZones(oracleKey string, triggers []TriggeredAbility) {
	if oracleKey == "" {
		return
	}
	for _, t := range triggers {
		if len(t.Zones) == 0 {
			continue
		}
		triggerZones.oracles[oracleKey] = true
		for _, zone := range t.Zones {
			if zone == ZoneBattlefield {
				continue
			}
			for _, kind := range t.Watches {
				triggerZones.addLocked(kind, zone)
			}
		}
	}
}

// addLocked records "an event of this kind must walk this zone".
// Named Locked for the file's convention only: the index is boot-time
// state, not game state.
func (ix *triggerZoneIndex) addLocked(kind EventKind, zone ZoneKind) {
	for _, z := range ix.zones[kind] {
		if z == zone {
			return
		}
	}
	ix.zones[kind] = append(ix.zones[kind], zone)
}

// zonesFor is the per-event question: which non-battlefield zones
// does this event kind have to be harvested from. Nil for almost
// every event of almost every game.
func (ix *triggerZoneIndex) zonesFor(kind EventKind) []ZoneKind {
	return ix.zones[kind]
}

// declares reports whether a catalog key has any non-battlefield
// trigger, so a walked zone can skip the catalog lookup for the
// cards that do not.
func (ix *triggerZoneIndex) declares(oracleKey string) bool {
	return oracleKey != "" && ix.oracles[oracleKey]
}

// harvestFromDeclaredZones is the ONE narrow harvest #925 adds: for
// each non-battlefield zone the index says this event kind is watched
// from, walk that zone and fire the abilities that declared it.
//
// Ordering: it runs after the battlefield walk, so a graveyard
// trigger queues behind a battlefield one that saw the same event.
// That is invisible to the rules — PendingTriggers is drained
// APNAP-style at the next priority boundary (CR 603.3b), not in
// harvest order — and it keeps the battlefield first, which is the
// order every existing test reads.
//
// Caller must hold g.mu in write mode.
func (g *Game) harvestFromDeclaredZones(pass *harvestPass) {
	if CatalogTriggers == nil {
		return
	}
	for _, kind := range triggerZones.zonesFor(pass.ev.Kind) {
		for _, z := range g.zonesOfKindLocked(kind) {
			g.harvestFromDeclaredZone(pass, z, kind)
		}
	}
}

// harvestFromDeclaredZone walks one zone for abilities that declared
// it. The source object is the card AS IT SITS IN THAT ZONE — which
// is the whole point: "when you cycle this card" is a graveyard
// card's ability, and the card in the graveyard is the object that
// has it (CR 400.7).
//
// CR 108.4: a card outside the battlefield and the stack has no
// controller, and its OWNER is the "you" of its abilities. So the
// source handed to AppliesTo / Build carries Controller = Owner,
// which is what makes `ByYou` mean "the owner's turn" and what
// decides who gets the resulting stack item. The same rule ADR 0062
// Decision 1 applies to activating from hand.
//
// Caller must hold g.mu in write mode.
func (g *Game) harvestFromDeclaredZone(pass *harvestPass, z *Zone, kind ZoneKind) {
	ev := pass.ev
	if z == nil {
		return
	}
	for i := range z.Cards {
		// The oracle key is the INDEX pre-filter's question — "does
		// any card with this key declare a zone at all" — and it is
		// answered before anything else because it is one map read
		// and it skips the whole of every ordinary card.
		if !triggerZones.declares(CatalogKey(z.Cards[i])) {
			continue
		}
		source := z.Cards[i]
		source.Controller = source.Owner
		// TriggersForCard, not CatalogTriggers: this is the fifth
		// place an OBJECT becomes triggered abilities, and it goes
		// through the same accessor the other four do (ADR 0071), so
		// a designation gate is evaluated here too. Off the
		// battlefield CR 400.7 has already cleared every designation,
		// so a Case's "Solved — …" is correctly not an ability it has
		// in a graveyard — which is exactly the answer a second,
		// parallel read here would have got wrong the first time a
		// card printed the combination.
		//
		// Behaviour-preserving for every card this walk exists for:
		// TriggersForCard reads CatalogAbilityKey, which degrades to
		// CatalogKey off the battlefield, and no declared zone is the
		// battlefield.
		triggers := TriggersForCard(source)
		if len(triggers) == 0 {
			continue
		}
		lki := source.Effective()
		for _, t := range triggers {
			if !TriggerWatchesFromZone(t, kind) {
				continue
			}
			if !triggerWatches(t.Watches, ev.Kind) {
				continue
			}
			if t.AppliesTo != nil && !t.AppliesTo(ev, &source, lki, g) {
				continue
			}
			g.harvestMatchLocked(pass, source, lki, t, false)
		}
	}
}
