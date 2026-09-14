package game

// ability_removal.go is the engine's ONE question for "what
// abilities does this object have right now" — CR 613.1f ability
// removal, made authoritative over the catalog.
//
// Why it exists. Every ability the catalog knows about is reached
// through a `Catalog*` hook keyed by oracle ID: `CatalogTriggers`
// for triggered abilities, `CatalogActivatedAbilities`,
// `CatalogManaAbilities`, `CatalogStaticAbilities`,
// `CatalogReplacements`, `CatalogCostModifiers`,
// `CatalogUntapStepPermissions`, `CatalogNoMaxHandSize`. Those hooks
// answer *what the printed card says*, which is the right question
// almost always and the wrong question under a layer-6 effect that
// takes the abilities away. Layer 6 could already clear
// `Characteristic.Abilities`, but that slice only carries keywords —
// so a Darksteel Mutation'd Sol Ring lost nothing it actually had,
// and a Song of the Dryads'd commander kept tapping for mana.
//
// Patching the readers would have fixed the readers. The defect is
// structural: the catalog was consulted at use time, from wherever
// the use happened, instead of through one accessor that could
// answer for the layers. So the check moves down here, and the
// readers above are reduced to naming which hook they want. The
// precedent is #539 (one zone-exit primitive replacing six
// hand-rolled movers) and #548 (one untap primitive replacing
// several), both taken for the same reason: a seventh caller must
// not be able to reintroduce the bug for free.
//
// This is ADR 0039's move applied to layer 6. That ADR made layer 4
// authoritative for the type predicates (`Card.IsCreature` reads
// `Effective().Types`, `PrintedIsCreature` is the other half); this
// one makes layer 6 authoritative for the ability lookups
// (`CatalogAbilityKey` reads the layered result, `CatalogKey` is the
// other half).
//
// # What ability removal removes, and what it does not
//
// CR 613.1f removes ABILITIES. It does not remove characteristics,
// and it does not remove anything that is not an ability at all. The
// boundary, stated once:
//
// Removed — activated abilities (including mana abilities and an
// Equipment's equip), triggered abilities, static abilities
// (anthems, cost modifiers, untap-step permissions, "you have no
// maximum hand size"), replacement effects, and keyword abilities.
//
// Kept — the object itself. It is still a permanent, still on the
// battlefield, still has a name (CR 613.1c is layer 3, a different
// layer), still has an owner and a controller, still carries its
// counters (CR 122 counters are objects on the permanent, not
// abilities — a planeswalker turned into a land by Song of the
// Dryads keeps its loyalty), still has its power and toughness
// (layer 7, applied after 6), still has whatever types layers 1-4
// gave it, and is still attached to whatever it was attached to and
// still has whatever is attached to IT. Ability removal is not a
// zone change and not an unattach.
//
// Kept too: every RESTRICTION on it. "Can't attack or block" is not
// an ability of the restricted creature — it is an effect the
// Pacifism has — so removing the creature's abilities does not lift
// it, and restrictions.go's separate `Characteristic.Restrictions`
// field is what makes that fall out rather than needing a rule here.
// The mirror case does go the other way, and through this file: an
// Aura whose OWN abilities are removed stops restricting anything,
// because a permanent with no abilities generates no continuous
// effect at all.
//
// Also kept: abilities an object has because of what it IS rather
// than because of what it says. CR 305.6's intrinsic mana ability of
// a land with a basic land type is granted by the land type, not
// printed in the rules text, so a permanent that a removal effect
// turned into a Forest taps for {G} — see ManaAbilitiesForCard,
// which gates the declared half and keeps the intrinsic half.
//
// # Timestamps (CR 613.6)
//
// An ability granted by an effect with a LATER timestamp than the
// removal still applies. Two halves, and they are handled in
// different places because they are different things:
//
//   - Granted keywords live in `Characteristic.Abilities`. The
//     layer-6 bucket is sorted by timestamp, an ability-removing
//     effect empties the slice in its own slot, and a grant sorted
//     after it appends to the emptied slice and survives. Rancor
//     cast after a Darksteel Mutation really does give the Insect
//     trample; cast before, it does not. That is the existing sort
//     doing the work — no new code.
//   - Printed abilities — everything the catalog answers for — are
//     part of the object, not granted to it, and a removal that
//     applies to the object removes them regardless of when either
//     arrived. That is why `AbilitiesRemoved` is one bool and not a
//     timestamp: there is no ordering question left to ask.
//
// The one ordering that IS asked is between removal effects
// themselves, and the layer-6 sort answers that too: see
// applyLayerLocked, which skips an effect whose source has already
// been silenced earlier in the same bucket.

// CatalogAbilityKey is the catalog key to use when asking what
// ABILITIES an object has right now. It is `CatalogKey` everywhere
// except on a permanent a layer-6 ability-removing effect applies
// to, where it is the empty string — and every catalog hook already
// treats an empty key as "this card has no entry", so a caller that
// switches to this accessor needs no other change.
//
// Use it for: triggered, activated, mana, static and replacement
// abilities, cost modifiers, untap-step permissions, and any future
// hook that answers "what does this permanent DO".
//
// Do NOT use it for: the target / enchant clause of a spell
// (`CatalogTargetSpec`), modes, additional and alternative costs,
// castable zones, "can't be countered", starting loyalty, battle
// defense, or the printed keyword list that seeds layer 0. None of
// those is an ability of a permanent on the battlefield; a spell
// being cast has not got a layered characteristic at all, and
// printed keywords are the baseline the removal is applied TO.
//
// Off the battlefield `Card.effective` is nil, so this degrades to
// `CatalogKey` exactly — which is CR 113.6 (a static ability does
// nothing while its source is elsewhere) falling out for free.
func CatalogAbilityKey(c Card) string {
	if c.HasLostAllAbilities() {
		return ""
	}
	return CatalogKey(c)
}

// HasLostAllAbilities reports whether a CR 613.1f ability-removing
// continuous effect currently applies to this object.
//
// Reads the layer engine's output directly rather than through
// `Effective()`, which returns a whole Characteristic by value — this
// is called once per battlefield card per trigger harvest and the
// copy is not free. The nil check IS the off-battlefield answer: a
// card in a hand or a graveyard has no effective characteristic and
// therefore no removal applied to it.
//
// Callers that hold a last-known-information snapshot instead of a
// live card (the LTB trigger harvest, the simultaneous-exit harvest)
// must read `Characteristic.AbilitiesRemoved` off that snapshot
// rather than calling this — the card has already left the
// battlefield and its cache has been cleared, but CR 603.10 says the
// trigger is judged on what the permanent looked like when it was
// still there.
func (c Card) HasLostAllAbilities() bool {
	return c.effective != nil && c.effective.AbilitiesRemoved
}
