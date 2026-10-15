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
//
// Both are read once, in CastSpell, where the permission is consumed,
// and copied out at once: the permission may be spent, swept or gone by
// the time the spell resolves. Enters-tapped then rides the spell
// (StackItem.EntersTapped) to its entry event, and the gains become one
// scoped effect registered against the spell's instance at cast, which
// applies the moment that instance is a permanent.

// permissionCastGains is what a consumed permission says about the
// permanent its cast produces.
type permissionCastGains struct {
	entersTapped bool
	keywords     []string
	abilities    []string
	label        string
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
		label:        p.SourceName,
	}
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
	if len(gains.keywords) == 0 && len(gains.abilities) == 0 {
		return
	}
	c, ok := g.LookupCardForEffect(cardID)
	if !ok || !c.IsPermanent() {
		return
	}
	var mods []Mod
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
