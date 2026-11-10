package game

import "github.com/google/uuid"

// cast_gains.go — the riders a cast permission puts on the permanent its
// cast produces (#3768, ADR 0066 amendment 2026-10-15).
//
// GrantsHaste (suspend, CR 702.62a) was the only one. Two printed
// sentences needed more, and both say "this way", which is what makes
// them properties of the PERMISSION rather than of the card:
//
//	"If you cast a spell this way, that artifact enters tapped."
//	    — Edgar, Master Machinist (CastPermission.EntersTapped)
//	"If you cast a creature spell this way, it gains haste and 'At the
//	 beginning of the end step, sacrifice this creature.'"
//	    — Strago and Relm (CastPermission.GainsKeywords / GainsAbilities)
//	"If you cast a creature spell this way, that creature enters with an
//	 additional +1/+1 counter on it."
//	    — Mikey & Don, Party Planners (CastPermission.EntersWithCounters, #4679)
//
// All are read once, in CastSpell, where the permission is consumed,
// and copied out at once: the permission may be spent, swept or gone by
// the time the spell resolves. Enters-tapped then rides the spell
// (StackItem.EntersTapped) to its entry event, the counter rides it as a
// StackItem.PromisedCounters mark, and the gains become one scoped
// effect registered against the spell's instance at cast, which applies
// the moment that instance is a permanent.

// permissionCastGains is what a consumed permission says about the
// permanent its cast produces.
type permissionCastGains struct {
	entersTapped bool
	keywords     []string
	abilities    []string
	subtypes     []string
	label        string
	// #4679: "that creature enters with an additional <kind> counter",
	// and the permission's source, which the mark names for the log.
	counterKind string
	counters    int
	source      uuid.UUID
	// anyPermanent widens the counter rider from creature spells to any
	// permanent spell (Noctis, Prince of Lucis: "that artifact enters with
	// a finality counter on it").
	anyPermanent bool
}

// permissionCastGainsOf copies the riders out of the permission the
// cast is claiming. A nil permission (a cast from hand, or one the
// card's own text allowed) says nothing.
func permissionCastGainsOf(p *CastPermission) permissionCastGains {
	if p == nil {
		return permissionCastGains{}
	}
	return permissionCastGains{
		entersTapped: p.EntersTapped,
		keywords:     append([]string(nil), p.GainsKeywords...),
		abilities:    append([]string(nil), p.GainsAbilities...),
		subtypes:     append([]string(nil), p.GainsSubtypes...),
		label:        p.SourceName,
		counterKind:  p.EntersWithCounterKind,
		counters:     p.EntersWithCounters,
		source:       p.Source,

		anyPermanent: p.EntersWithCountersPermanent,
	}
}

// markPermissionEntryCountersLocked writes the permission's "that
// creature enters with an additional counter" rider (#4679) onto the
// spell `cardID` that the cast just put on the stack, through the one
// writer the next-spell promise and Wildgrowth Archaic's cast trigger
// use (addPromisedCounterLocked). From there it is the existing read
// side: applyCastEntryCountersLocked seeds it onto the entry event
// before the CR 614 pipeline (CR 614.1c, CR 614.12), so replacements
// that modify counters (Doubling Season) see it.
//
// Nothing for a spell that is not a creature spell as cast (a permanent
// spell, when the permission says EntersWithCountersPermanent): the card on
// the stack has its cast face and bestow status by now, so an
// Adventure, a modal card's noncreature face and a bestowed Aura spell
// all read as what they are (CR 702.103b). The mark is on the item, so
// it goes wherever the spell goes: a countered spell takes it with it
// (CR 400.7) and a copy made of the spell does not have it (CR 707.2).
// A copy that is itself CAST through a permission (CR 707.12) is
// marked, but a permanent spell's copy resolves into a token that reads
// none of the cast riders, EntersTapped included; no permission with a
// rider casts a copy today.
//
// Caller must hold g.mu (write).
func (g *Game) markPermissionEntryCountersLocked(cardID uuid.UUID, gains permissionCastGains) {
	if gains.counterKind == "" || gains.counters < 1 {
		return
	}
	spell, item := g.spellCardOnStackLocked(cardID), g.StackMeta[cardID]
	if spell == nil || item == nil || !(spell.IsCreature() || (gains.anyPermanent && spell.IsPermanent())) {
		return
	}
	addPromisedCounterLocked(item, PromisedCounter{
		Kind: gains.counterKind, Count: gains.counters, Source: gains.source, SourceName: gains.label,
	})
}

// grantGainsForCastLocked registers "it gains <keywords> and
// '<abilities>'" for the permanent the spell `cardID` will become: one
// layer-6 scoped effect with AddKeywordsMod and GrantAbilitiesMod,
// pinned to the instance with no controller (an effect with no duration
// outlives a control change), lasting for
// ForThePermanentThisSpellBecomesDuration. Nothing for a spell that is
// not a permanent spell (CR 110.1 — an instant or sorcery never becomes
// an object that could gain anything), and nothing when there is
// nothing to gain.
//
// Caller must hold g.mu (write).
func (g *Game) grantGainsForCastLocked(cardID uuid.UUID, gains permissionCastGains) {
	if len(gains.keywords) == 0 && len(gains.abilities) == 0 && len(gains.subtypes) == 0 {
		return
	}
	c, ok := g.LookupCardForEffect(cardID)
	if !ok || !c.IsPermanent() {
		return
	}
	var mods []Mod
	if len(gains.subtypes) > 0 {
		mods = append(mods, AddSubtypesMod(gains.subtypes...))
	}
	if len(gains.keywords) > 0 {
		mods = append(mods, AddKeywordsMod(gains.keywords...))
	}
	if len(gains.abilities) > 0 {
		mods = append(mods, GrantAbilitiesMod(gains.abilities...))
	}
	label := "Cast this way — gains"
	if gains.label != "" {
		label = gains.label + " — cast this way, gains"
	}
	g.RegisterScopedEffectForEffect(cardID, []AffectedObject{{ID: cardID}}, mods,
		ForThePermanentThisSpellBecomesDuration(cardID), label)
}
