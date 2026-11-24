package game

import "github.com/google/uuid"

// spell_grants.go — "it gains '<ability>'" given to a PERMANENT spell on
// the stack (#6088): Mizzix, Replica Rider's copy "gains haste and 'At the
// beginning of your end step, sacrifice this permanent.'"
//
// The grant is a ScopedEffect record PINNED to the spell (PinStackObject)
// with a grantAbilities mod, as a keyword grant to a spell is
// (spell_keywords.go): data, ended with the object on the stack unless it
// becomes a permanent, to which repinSpellControlLocked hands it (CR
// 400.7a) for the rest of its duration. A copy of a permanent spell
// becomes a token with a new ID (CR 707.10f), which is why the pin, and
// not an ID record, is what carries it there.
//
// The stack step does not apply the bundle: the abilities of a permanent
// spell function on the battlefield (CR 113.6), and every bundle the
// catalog grants this way is a permanent's ability. An instant or sorcery
// spell is refused, so nothing is silently dropped on one.

// GrantAbilitiesToPermanentSpellForEffect gives the permanent spell
// `spellID` on the stack the catalog ability bundles named (ADR 0093) for
// the duration `d`, carried to the permanent it becomes (CR 400.7a).
// Reports false, registering nothing, when `spellID` is not a permanent
// spell on the stack or there is nothing to grant.
//
// Caller must hold g.mu (write).
func (g *Game) GrantAbilitiesToPermanentSpellForEffect(sourceID, spellID uuid.UUID, grants []string, d Duration, label string) bool {
	if len(grants) == 0 {
		return false
	}
	card, _, ok := g.stolenSpellLocked(spellID)
	if !ok || !card.IsPermanent() {
		return false
	}
	if !g.registerScopedEffectLocked(sourceID,
		[]AffectedObject{PinStackObject(spellID, card.ObjectEpoch)},
		[]Mod{GrantAbilitiesMod(grants...)},
		g.PinnedToStack(d, spellID), label, timeNowUnixNano()) {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	return true
}
