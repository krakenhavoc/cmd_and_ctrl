package game

import (
	"sort"

	"github.com/google/uuid"
)

// spell_keywords.go — layer 6 for SPELLS (ADR 0107 §3 decision 5, owner
// decision 3, #1854).
//
//	CR 613.1f  "Layer 6: Ability-adding effects, keyword counters,
//	            ability-removing effects, and effects that say an
//	            object can't have an ability are applied."
//
// A spell is an object (CR 109.1), and "that spell gains rebound"
// (Taigam, Ojutai Master) and "Instant and sorcery spells you control
// have rebound" (Cast Through Time) are ability-adding effects on it.
// Until this file the stack step of the layer pass (ADR 0104,
// stackControlPassLocked) applied layer 2 only, so a spell could change
// hands and could not gain anything.
//
// The step is now layers 2 and 6, keywords only. For each spell, once
// its controller is settled (layer 2 before layer 6, CR 613.1), two
// sources contribute, applied in timestamp order (CR 613.7) as the
// battlefield pass applies its own:
//
//  1. A ScopedEffect record PINNED to the spell (PinStackObject) with
//     an addKeywords mod: an effect of a resolving spell or ability
//     (CR 611.2). GrantKeywordsToSpellForEffect writes one. It is data,
//     so a table with a spell given rebound is a restore point, and its
//     stack pin ends with the object (CR 400.7).
//  2. A battlefield static declared StaticAbility.AffectsSpells: a
//     static ability, so its set is never locked in (CR 611.3a). A
//     spell cast after Cast Through Time entered has rebound, and every
//     spell loses it the moment Cast Through Time leaves.
//
// The answer is materialised onto Card.stackGranted, which HasKeyword
// and Effective() read beside the printed keywords. That is the whole
// reason rebound.go needed no change: spellRebounds asks HasKeyword of
// the resolving spell, and a granted rebound is in the list it reads.
//
// Only ADDING keywords is admitted on a stack pin (stackPinProblem).
// Nothing in the catalog removes an ability from a spell, sets one's
// types or colours, or changes its power, so the stack step does not
// pretend to.

// spellStaticGrant is one battlefield static over spells, bound to the
// permanent it comes from.
type spellStaticGrant struct {
	ability StaticAbility
	source  *Card
	ts      int64
}

// spellStaticsLocked gathers every battlefield static that affects
// spells, once per pass. A source whose abilities the battlefield pass
// just removed (CR 613.1f, Humility and its kin) contributes nothing:
// the battlefield pass has run, so HasLostAllAbilities is its final
// answer for this pass.
//
// Caller must hold g.mu (write); the source pointers are into the
// battlefield slice and are good only for this pass.
func (g *Game) spellStaticsLocked() []spellStaticGrant {
	if g.Battlefield == nil || CatalogStaticAbilities == nil {
		return nil
	}
	var out []spellStaticGrant
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if src.HasLostAllAbilities() {
			continue
		}
		for _, ab := range StaticAbilitiesForCard(*src) {
			if ab.AffectsSpells && ab.AppliesTo != nil {
				out = append(out, spellStaticGrant{ability: ab, source: src, ts: src.layerTimestamp()})
			}
		}
	}
	return out
}

// spellKeywordContribution is one effect's layer-6 contribution to a
// spell, for the timestamp sort.
type spellKeywordContribution struct {
	ts    int64
	apply func(ch *Characteristic)
}

// spellGrantedKeywordsLocked is layer 6 for the spell `c`: the keywords
// the pinned records and the spell statics give it, in timestamp order.
// Nil when nothing gives it anything, which is every spell at nearly
// every table.
//
// Caller must hold g.mu (write).
func (g *Game) spellGrantedKeywordsLocked(c *Card, statics []spellStaticGrant) []string {
	var parts []spellKeywordContribution
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if !pinsStackObject(e.Affected, c.InstanceID, c.ObjectEpoch) {
			continue
		}
		for _, m := range e.Mods {
			if m.Kind != ModAddKeywords {
				continue
			}
			keywords := m.Keywords
			parts = append(parts, spellKeywordContribution{ts: e.Timestamp, apply: func(ch *Characteristic) {
				for _, kw := range keywords {
					ch.Abilities = AppendKeywordAbility(ch.Abilities, kw)
				}
			}})
		}
	}
	for _, s := range statics {
		if !s.ability.AppliesTo(c, g, s.source) {
			continue
		}
		s := s
		parts = append(parts, spellKeywordContribution{ts: s.ts, apply: func(ch *Characteristic) {
			if s.ability.Apply != nil {
				s.ability.Apply(ch, c, g, s.source)
			}
		}})
	}
	if len(parts) == 0 {
		return nil
	}
	sort.SliceStable(parts, func(i, j int) bool { return parts[i].ts < parts[j].ts })
	var ch Characteristic
	for _, p := range parts {
		p.apply(&ch)
	}
	return ch.Abilities
}

// stackKeywordPassLocked is the layer-6 half of the stack step: it
// stamps every spell's granted keywords onto its stack card. Called by
// stackControlPassLocked once control is settled.
//
// A spell mid-resolution has no stack item and is skipped, like the
// layer-2 half: its card was copied out before the item went
// (resolveTopOfStackLocked), so the resolution reads the keywords it
// had as it began to resolve.
//
// Caller must hold g.mu (write).
func (g *Game) stackKeywordPassLocked() {
	if g.Stack == nil || len(g.Stack.Cards) == 0 {
		return
	}
	statics := g.spellStaticsLocked()
	for i := range g.Stack.Cards {
		c := &g.Stack.Cards[i]
		item := g.StackMeta[c.InstanceID]
		if item == nil || item.Kind != StackItemSpell {
			continue
		}
		c.stackGranted = g.spellGrantedKeywordsLocked(c, statics)
	}
}

// spellStaticIsLiveLocked reports whether any permanent on the
// battlefield has a static over spells. The gate on the EventCast bump
// in layerVersionBump: a spell cast while Cast Through Time is out has
// rebound from the moment it is on the stack, and a cast moves nothing
// on the battlefield.
//
// Caller must hold g.mu.
func spellStaticIsLiveLocked(g *Game) bool {
	return staticOnBattlefieldLocked(g, func(ab StaticAbility) bool { return ab.AffectsSpells })
}

// GrantKeywordsToSpellForEffect gives the spell `spellID` on the stack
// the keywords named, for the duration `d`, and for no longer than it is
// that object on the stack — "that spell gains rebound" (Taigam, Ojutai
// Master; Ojer Pakpatiq, Deepest Epoch), "it gains sunburst" (Lux
// Artillery), "it gains haste until end of turn" (Tyvar Kell's emblem,
// Generator Servant's mana). An ability-adding effect (CR 611.2,
// 613.1f), written as a ScopedEffect pinned to the spell: data, so it
// survives a restore point, and it ends when the object leaves the stack
// (CR 400.7) or when `d` does, whichever is first.
//
// A permanent spell hands it to the permanent it becomes (CR 400.7a,
// inheritSpellControlLocked), and the duration goes with it (ADR 0109
// §11 decision 3): "haste until end of turn" on the spell is haste until
// end of turn on the creature, and an indefinite grant lasts as long as
// the permanent does. Pass IndefiniteDuration() for a grant that names
// no duration.
//
// Reports false, registering nothing, when `spellID` is not a spell on
// the stack — it resolved or was countered before the granting ability
// resolved — or there is nothing to grant.
//
// It recomputes before it returns, so the spell has the keywords at
// once.
//
// Caller must hold g.mu (write). Effects call it from inside the
// resolution frame, which already holds it.
func (g *Game) GrantKeywordsToSpellForEffect(sourceID, spellID uuid.UUID, keywords []string, d Duration, label string) bool {
	if len(keywords) == 0 {
		return false
	}
	card, _, ok := g.stolenSpellLocked(spellID)
	if !ok {
		return false
	}
	if !g.registerScopedEffectLocked(sourceID,
		[]AffectedObject{PinStackObject(spellID, card.ObjectEpoch)},
		[]Mod{AddKeywordsMod(keywords...)},
		g.PinnedToStack(d, spellID), label, timeNowUnixNano()) {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	return true
}
