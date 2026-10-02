package game

import "github.com/google/uuid"

// unpreventable_damage.go — every "damage can't be prevented" in the
// engine, read at one gate (ADR 0107 §5, #1853), and the redirection
// ban two of those cards print beside it.
//
// THE RULE. CR 615.12: "If unpreventable damage would be dealt, any
// applicable prevention effects are still applied to it. Those effects
// won't prevent any damage, but any additional effects they have will
// take place. Existing damage prevention shields won't be reduced by
// damage that can't be prevented." CR 615.12a: a prevention effect is
// applied to an unpreventable damage event only once.
//
// NOT A KEYWORD, AND NOT A LAYER EFFECT. "Damage can't be prevented" is
// a rule-modifying effect (CR 613.11), the same kind of effect as "can't
// be countered" (cant_be_countered.go, ADR 0106 §4), so it is read at
// one moment: while the CR 614 window settles a damage event. A static
// is read live there (CR 611.3a); a resolved spell's "this turn" covers
// damage dealt later in the turn (CR 611.2c) and ends at cleanup
// (CR 514.2).
//
// THE FOUR SOURCES (ADR 0107 §5 decision 2), asked in this order, the
// first yes winning:
//
//  1. The damage instruction's own mark — "the damage can't be
//     prevented" (Combust, Pinpoint Avalanche, Banefire with X of 5 or
//     more): DamageMarks.CantBePrevented on the event's tail, stamped by
//     DealMarkedDamageForEffect.
//  2. The source's own static — "damage that would be dealt by this
//     creature can't be prevented" (Excruciator, Malignus): read once
//     as the event is opened, from the source's last-known information
//     once it has left (damageTail.sourceUnpreventable).
//  3. Battlefield statics — "Damage can't be prevented" (Leyline of
//     Punishment, Sunspine Lynx), "Combat damage can't be prevented"
//     (Frenzied Baloth) and "Combat damage that would be dealt by
//     creatures you control can't be prevented" (Questing Beast),
//     read off the battlefield on every ask and keyed by
//     CatalogAbilityKey, so a Leyline that has lost its abilities
//     (CR 613.1f) stops.
//  4. The rule grants — a ModDamageCantBePrevented ScopedEffect:
//     ScopeGame for "damage can't be prevented this turn" (Skullcrack,
//     Stomp), pinned to one permanent for "damage that would be dealt
//     to that creature this turn can't be prevented" (Whippoorwill).
//
// WHAT THE APPLY-LOOP DOES WITH THE ANSWER (decision 3). A replacement
// that declared itself a prevention effect (ReplacementEffect.Prevention)
// is settled before any CR 616 prompt is built: marked applied for the
// event, once, with its Replace not run. So the damage is dealt in full,
// a charged shield keeps its charge, and protection does not stop it
// (CR 702.16e is a prevention effect). A replacement that is not a
// prevention effect — a doubler, a redirection — is untouched.
//
// "Any additional effects they have will take place": ADR 0107 §6's
// "the damage prevented this way" follow-up (ModPreventNextFromSource's
// Then) is owed here with zero prevented, and runs once for the instance,
// as it would have after a real prevention
// (preventionAppliedToUnpreventableLocked).
//
// CR 615.13's "whenever damage is prevented" triggers have nothing to
// see either: nothing was prevented, and the engine emits no prevention
// event to trigger on.

// DamageMarks are the riders a damage INSTRUCTION can carry about its
// own damage. The zero value is ordinary damage.
type DamageMarks struct {
	// CantBePrevented is "the damage can't be prevented" (CR 615.12).
	CantBePrevented bool
	// CantBeRedirected is Lava Burst's "or dealt instead to another
	// permanent or player": no redirection replacement applies to it.
	CantBeRedirected bool
}

// SpellCondition is a printed condition on a rider of a SPELL's own,
// judged while the spell is on the stack (or resolving): Banefire's "if X
// is 5 or more", Demonfire's hellbent, a kicked Urza's Rage. `item` is the
// spell's stack item. Catalog data; nil means the rider never applies.
type SpellCondition func(g *Game, item *StackItem) bool

// CatalogSpellDamageUnpreventable returns the condition under which a
// spell with the given catalog key says "the damage can't be prevented"
// of its own damage (Combust, Pinpoint Avalanche, Banefire with X of 5 or
// more), or nil. carddef.go sets it from CardDef.SpellDamageCantBePrevented.
var CatalogSpellDamageUnpreventable func(key string) SpellCondition

// spellDamageUnpreventableLocked is source 1 for a spell: `sourceID` is a
// spell on the stack, or resolving, whose own text says its damage can't
// be prevented, and its condition holds now. Read through the same
// lookup the damage tail uses, so a copy (CR 707.10) carries its
// original's rider and its copied X.
//
// Caller must hold g.mu.
func (g *Game) spellDamageUnpreventableLocked(sourceID uuid.UUID) bool {
	if CatalogSpellDamageUnpreventable == nil || sourceID == uuid.Nil {
		return false
	}
	card, item, ok := g.stackSpellLocked(sourceID)
	if !ok || item == nil || item.Kind != StackItemSpell {
		return false
	}
	key := CatalogKey(card)
	if key == "" {
		return false
	}
	cond := CatalogSpellDamageUnpreventable(key)
	return cond != nil && cond(g, item)
}

// SpellDamageCantBePreventedForEffect is spellDamageUnpreventableLocked
// for the view projection's stack chip (ADR 0107 §5 decision 5): the
// table is told before a Fog finds out the hard way. Caller must hold
// g.mu (read or write).
func (g *Game) SpellDamageCantBePreventedForEffect(id uuid.UUID) bool {
	return g.spellDamageUnpreventableLocked(id)
}

// UnpreventableDamageScope is which damage a printed "can't be
// prevented" static covers.
type UnpreventableDamageScope uint8

const (
	// UnpreventableByThis is "Damage that would be dealt by this
	// creature can't be prevented" (Excruciator, Malignus): the damage
	// whose source is the permanent with the static.
	UnpreventableByThis UnpreventableDamageScope = iota
	// UnpreventableAll is "Damage can't be prevented" (Leyline of
	// Punishment, Sunspine Lynx): every damage event.
	UnpreventableAll
	// UnpreventableCombat is "Combat damage can't be prevented"
	// (Frenzied Baloth).
	UnpreventableCombat
	// UnpreventableCombatByYourCreatures is "Combat damage that would be
	// dealt by creatures you control can't be prevented" (Questing
	// Beast): combat damage whose source, as it was when the damage was
	// dealt, is a creature controlled by the static's controller.
	UnpreventableCombatByYourCreatures
)

// UnpreventableDamageStatic is one printed "damage can't be prevented"
// static on a permanent. Catalog data, never stored: the gate reads it
// off the battlefield on every ask.
type UnpreventableDamageStatic struct {
	// Label is the clause as printed.
	Label string
	// Scope is which damage it covers.
	Scope UnpreventableDamageScope
}

// CatalogUnpreventableDamage returns the "damage can't be prevented"
// statics a permanent with the given catalog key has. carddef.go sets it
// from CardDef.DamageCantBePrevented; a game-package test may stub it.
var CatalogUnpreventableDamage func(key string) []UnpreventableDamageStatic

// unpreventableStaticsOf is the statics a permanent has right now: its
// catalog entry's, keyed by CatalogAbilityKey so one that has lost its
// abilities has none (CR 613.1f).
func unpreventableStaticsOf(c Card) []UnpreventableDamageStatic {
	if CatalogUnpreventableDamage == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return CatalogUnpreventableDamage(key)
}

// damageUnpreventableLocked reports whether the damage event can't be
// prevented, asking the four sources in the order the file comment
// lists.
//
// Caller must hold g.mu. Reads only.
func (g *Game) damageUnpreventableLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventDamage {
		return false
	}
	// 1 and 2: the instruction's mark and the source's own static.
	if t := ev.damageTail; t != nil && (t.marks.CantBePrevented || t.sourceUnpreventable) {
		return true
	}
	// 1, for a spell: its own printed "the damage can't be prevented",
	// judged as it deals the damage.
	if g.spellDamageUnpreventableLocked(ev.DamageSource) {
		return true
	}
	// 3: the battlefield statics.
	if g.unpreventableOnBattlefieldLocked(ev) {
		return true
	}
	// 4: the rule grants.
	return g.scopedRuleCoversDamageLocked(ev, ModDamageCantBePrevented)
}

// damageCantBeRedirectedLocked reports whether no redirection may apply
// to the damage event: Lava Burst's mark on its own damage to a
// creature, or Whippoorwill's grant on the creature being dealt it.
//
// Caller must hold g.mu. Reads only.
func (g *Game) damageCantBeRedirectedLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventDamage {
		return false
	}
	if t := ev.damageTail; t != nil && t.marks.CantBeRedirected {
		return true
	}
	return g.scopedRuleCoversDamageLocked(ev, ModDamageCantBeRedirected)
}

// unpreventableOnBattlefieldLocked is source 3: some permanent's static
// makes this damage unpreventable (CR 604.2: active while the permanent
// is on the battlefield and has the ability). A phased-out permanent is
// not in the battlefield slice (CR 702.26b).
//
// Caller must hold g.mu.
func (g *Game) unpreventableOnBattlefieldLocked(ev *ReplacementEvent) bool {
	if CatalogUnpreventableDamage == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		for _, s := range unpreventableStaticsOf(*src) {
			switch s.Scope {
			case UnpreventableAll:
				return true
			case UnpreventableCombat:
				if ev.IsCombatDamage {
					return true
				}
			case UnpreventableCombatByYourCreatures:
				// The source as it was when the damage was dealt
				// (CR 608.2h): a creature its controller controlled.
				lki := ev.SourceLKI
				if ev.IsCombatDamage && lki != nil && lki.Controller == src.Controller && characteristicIsCreature(lki) {
					return true
				}
			}
		}
	}
	return false
}

// characteristicIsCreature reports whether a characteristic snapshot is
// a creature's.
func characteristicIsCreature(ch *Characteristic) bool {
	for _, t := range ch.Types {
		if t == "Creature" {
			return true
		}
	}
	return false
}

// sourceDamageCantBePreventedLocked is source 2, read once as a damage
// event is opened: does the source's own static say its damage can't be
// prevented (UnpreventableByThis)?
//
// A live permanent is read through CatalogAbilityKey. A source that has
// left the battlefield is judged as it last existed there (CR 608.2h) —
// the departed object's catalog entry, unless it had lost its abilities
// (Characteristic.AbilitiesRemoved). A spell, an emblem or an unknown
// source has no such static: the ability functions only on the
// battlefield.
//
// Caller must hold g.mu.
func (g *Game) sourceDamageCantBePreventedLocked(sourceID uuid.UUID) bool {
	if sourceID == uuid.Nil || CatalogUnpreventableDamage == nil {
		return false
	}
	byThis := func(statics []UnpreventableDamageStatic) bool {
		for _, s := range statics {
			if s.Scope == UnpreventableByThis {
				return true
			}
		}
		return false
	}
	if src := findBattlefieldCard(g, sourceID); src != nil {
		return byThis(unpreventableStaticsOf(*src))
	}
	rec, ok := g.departedDamageSourceLocked(sourceID, nil)
	if !ok || rec.Characteristic.AbilitiesRemoved {
		return false
	}
	card, ok := g.LookupCardForEffect(sourceID)
	if !ok {
		return false
	}
	key := ownAbilityKey(card)
	if key == "" {
		return false
	}
	return byThis(CatalogUnpreventableDamage(key))
}

// scopedRuleCoversDamageLocked is source 4 for `kind`: a live record of
// that kind covers this damage — every damage event for ScopeGame, the
// damage dealt TO a pinned permanent otherwise.
//
// Caller must hold g.mu.
func (g *Game) scopedRuleCoversDamageLocked(ev *ReplacementEvent, kind ModKind) bool {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if !scopedEffectHasMod(e, kind) {
			continue
		}
		if e.Scope == ScopeGame {
			return true
		}
		if e.Scope == ScopeNone && scopedAffectsLiveObjectLocked(g, *e, ev.DamageTarget) {
			return true
		}
	}
	return false
}

// settleUnpreventableLocked is the apply-loop's CR 615.12 step, run on
// every gathered list before anything else: a prevention effect gathered
// for damage that can't be prevented — or a redirection for damage that
// can't be redirected — is marked applied for this event (once, CR
// 615.12a) without its Replace running, and dropped from the list. The
// rest is returned in gather order.
//
// Caller must hold g.mu, and must have allocated the once-per-event map
// entry for ev.ID.
func (g *Game) settleUnpreventableLocked(ev *ReplacementEvent, applicable []activeReplacement) []activeReplacement {
	if ev == nil || ev.Kind != RepEventDamage || len(applicable) == 0 {
		return applicable
	}
	var prevents, redirects bool
	for _, a := range applicable {
		prevents = prevents || a.effect.Prevention
		redirects = redirects || a.effect.RedirectsDamage
	}
	if !prevents && !redirects {
		return applicable
	}
	unpreventable := prevents && g.damageUnpreventableLocked(ev)
	noRedirect := redirects && g.damageCantBeRedirectedLocked(ev)
	if !unpreventable && !noRedirect {
		return applicable
	}
	out := make([]activeReplacement, 0, len(applicable))
	for _, a := range applicable {
		if (unpreventable && a.effect.Prevention) || (noRedirect && a.effect.RedirectsDamage) {
			g.replacementsAppliedThisEvent[ev.ID][a.id] = true
			g.preventionAppliedToUnpreventableLocked(ev, a)
			continue
		}
		out = append(out, a)
	}
	return out
}

// preventionAppliedToUnpreventableLocked is where CR 615.12's "any
// additional effects they have will take place" happens, with nothing
// prevented. The one prevention effect with an additional effect is ADR
// 0107 §6's next-damage shield, whose "the damage prevented this way"
// follow-up is queued with zero; the shield itself is not used up
// (CR 609.7b: a shield that prevents no damage isn't).
//
// Caller must hold g.mu (write).
func (g *Game) preventionAppliedToUnpreventableLocked(ev *ReplacementEvent, a activeReplacement) {
	g.preventionFollowUpForUnpreventableLocked(ev, a.id)
}

// --- the rule grants (source 4) -------------------------------------

// DamageCantBePreventedThisTurnForEffect is "damage can't be prevented
// this turn" (Skullcrack, Stomp, Flaring Pain; CR 615.12): every damage
// event until cleanup (CR 514.2). It modifies the rules rather than any
// object, so it covers damage from a source that did not exist when it
// began (CR 611.2c). Reports whether a record was written.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) DamageCantBePreventedThisTurnForEffect(sourceID uuid.UUID, label string) bool {
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModDamageCantBePrevented}}, g.UntilEndOfTurnDuration(), label)
}

// DamageToCantBePreventedThisTurnForEffect is "damage that would be
// dealt to that creature this turn can't be prevented" — with
// noRedirect, "… or dealt instead to another permanent or player"
// (Whippoorwill). The record is pinned to the permanent (CR 400.7), so
// it ends if the creature leaves. Registers nothing for a card that is
// not on the battlefield.
//
// Caller must hold g.mu (write).
func (g *Game) DamageToCantBePreventedThisTurnForEffect(sourceID, cardID uuid.UUID, noRedirect bool, label string) bool {
	affected := g.PinnedObjectsLocked(cardID)
	if len(affected) == 0 {
		return false
	}
	mods := []Mod{{Kind: ModDamageCantBePrevented}}
	if noRedirect {
		mods = append(mods, Mod{Kind: ModDamageCantBeRedirected})
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, uuid.Nil, mods,
		g.PinnedTo(g.UntilEndOfTurnDuration(), cardID), label, timeNowUnixNano())
}

// DamageCantBePreventedThisTurnLabels lists the live game-wide "damage
// can't be prevented this turn" grants by label, oldest first — the game
// banner's line (ADR 0107 §5 decision 5). Empty on nearly every turn.
//
// Caller must hold g.mu (read or write).
func (g *Game) DamageCantBePreventedThisTurnLabels() []string {
	var out []string
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Scope != ScopeGame || !scopedEffectHasMod(e, ModDamageCantBePrevented) {
			continue
		}
		label := e.SourceName
		if label == "" {
			label = e.Label
		}
		out = append(out, label)
	}
	return out
}
