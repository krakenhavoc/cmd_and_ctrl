package game

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// redirect_damage.go — ADR 0108 §9 (#1905): damage dealt to something
// else instead.
//
// THE RULES.
//
//   - CR 614.9: "Some effects replace damage dealt to one battle,
//     creature, planeswalker, or player with the same damage dealt to
//     another battle, creature, planeswalker, or player; such effects are
//     called redirection effects. If one of those permanents is no longer
//     on the battlefield when the damage would be redirected, or is no
//     longer a battle, creature, or planeswalker when the damage would be
//     redirected, the effect does nothing. If damage would be redirected
//     to or from a player who has left the game, the effect does nothing."
//   - CR 614.5: a replacement effect gets one opportunity to affect an
//     event "or any modified events that may replace that event", so a
//     redirection never applies to its own result, and the part of an
//     event a charged redirection splits off keeps every effect already
//     applied to the event as applied.
//   - CR 616.1f, 616.2: the modified event is gathered again, so effects
//     that now apply — a shield on the new recipient, its protection —
//     apply to it, and the shields of the old recipient no longer do.
//   - CR 120.4b: damage is dealt "as modified by replacement and
//     prevention effects", and its results are processed after (120.4c),
//     so a redirected source keeps its deathtouch, lifelink, infect and
//     wither, combat damage stays combat damage, and a commander's
//     redirected combat damage counts toward the new recipient's
//     CR 903.10a total.
//   - CR 120.8: a source that would deal 0 damage deals none, so there is
//     nothing to redirect and a "next time" redirection is not spent.
//   - CR 609.7b: a shield that replaces no damage is not used up — so a
//     redirection that would do nothing (CR 614.9) does not apply at all.
//   - CR 615.7, the Harm's Way ruling: "If the chosen source would
//     simultaneously deal damage to multiple permanents you control …
//     Harm's Way will redirect just 2 of that damage. … You choose which 2
//     damage is redirected" — divide_shield.go, owner decision 1.
//
// "CAN'T BE PREVENTED" IS NOT "CAN'T BE REDIRECTED". A redirection is a
// replacement effect (CR 614.1a), not a prevention effect (CR 615.1a), so
// CR 615.12 leaves it alone: damage that can't be prevented is still dealt
// to Beacon of Destiny instead. What stops it is "that damage can't be …
// dealt instead to another permanent or player" (Lava Burst's mark,
// Whippoorwill's grant): every redirection declares
// ReplacementEffect.RedirectsDamage, and settleUnpreventableLocked applies
// it without running it under damageCantBeRedirectedLocked, exactly as a
// prevention effect is under CR 615.12 — so a "next time" redirection is
// not spent and a charged one keeps its charge (CR 609.7b).
//
// THE PRIMITIVE (§9 decision 1). redirectDamageEventLocked rewrites the
// event's recipient and its tail's kind together, so redirected damage
// lands through the player path or the permanent path it is now dealt to.
// Every redirection uses it, scoped or static (RedirectDamageEventForEffect
// is its catalog face), and TestDamageReplacementsDeclareWhetherTheyPrevent
// fails a Replace that writes DamageTarget by hand.
//
// A CHARGED REDIRECTION THAT COVERS PART OF AN EVENT. "The next 2 damage …
// is dealt to any target instead" meeting 5 damage redirects 2 and leaves
// 3 where it was. The event keeps the 3 and goes on through the window;
// the 2 rides its tail (damageTail.redirected) as a part, and is dealt as
// its own event when the event reaches its terminal outcome, with every
// effect already applied to the event marked applied to it (CR 614.5).
//
// THE KIND (§9 decision 2). ModRedirectDamage, plain data in a ScopedEffect
// record, reads:
//
//   - the source, as every damage kind does (Objects + SourceZone, read by
//     damageFromChosenSourceLocked; Queries, rechecked as the damage would
//     be dealt, CR 609.7b), or nothing ("the next time damage would be
//     dealt to this creature");
//   - what it protects: Player, with Types for "you and/or permanents you
//     control", a pinned permanent, or nothing ("would deal damage this
//     turn", Opal-Eye and Reflect Damage);
//   - how long: Next (one instance, spent through SpentInstance as a
//     next-damage shield is), Amount (CR 615.7's charge, divided by
//     divide_shield), or neither (all this turn: Kor Chant, Oracle's
//     Attendants); CombatOnly narrows it to combat damage;
//   - where to: To (one object, epoch-pinned) or ToSourceController;
//   - Then, Eye for an Eye's "instead that source deals that much damage to
//     you and Eye for an Eye deals that much damage to that source's
//     controller": a record with no destination leaves the event as it is
//     and owes Then with the amount, through the one follow-up queue
//     (PreventionFollowUp). It does not declare RedirectsDamage: nothing
//     is dealt instead to another permanent or player (§9 decision 3).

// ModRedirectDamage is "<damage> is dealt to <destination> instead" (ADR
// 0108 §9, CR 614.9). Reads Objects, SourceZone, Queries, Player, Types,
// CombatOnly, Next, Amount, To, ToSourceController, Then, SpentBatch and
// SpentInstance. Scope ScopeGame for a protected player or none; pinned
// (ScopeNone) to a protected permanent.
const ModRedirectDamage ModKind = "redirectDamage"

// DamageRedirection is the queue-side description of a ModRedirectDamage
// record. The source and protection fields mean what DamageShield's do.
type DamageRedirection struct {
	// EffectSource is the card whose spell or ability made the effect;
	// Controller is its "you", the controller of any follow-up.
	EffectSource uuid.UUID
	Controller   uuid.UUID

	// Source is the chosen or targeted source (DamageSourceRefLocked);
	// the zero ref is any source matching Queries, or any source at all.
	Source     ObjectRef
	SourceZone ZoneKind
	Queries    []PermanentQuery

	// ProtectPlayer, ProtectTypes and ProtectPermanent are whose damage is
	// redirected; all zero is anyone's. ProtectPlayer beside
	// ProtectPermanent is "this creature and/or you" (Glarecaster).
	// ProtectRecipients is "to an opponent" (Soltari Guerrillas), the
	// multiplier's vocabulary read against Controller, and stands alone.
	ProtectPlayer     uuid.UUID
	ProtectTypes      []string
	ProtectPermanent  uuid.UUID
	ProtectRecipients DamageRecipients

	// CombatOnly is "combat damage".
	CombatOnly bool

	// Next is "the next time" (CR 615.8's instance); Amount is "the next
	// N damage" (CR 615.7's charge). Neither is all damage this turn.
	Next   bool
	Amount int

	// To is the player or permanent the damage is dealt to instead,
	// pinned to the object it is now (CR 400.7). ToSourceController is
	// "that source's controller" instead. Neither, with Then, is Eye for
	// an Eye's shape.
	To                 uuid.UUID
	ToSourceController bool

	// Then is a follow-up body owed with the amount (Eye for an Eye).
	Then BodyRef

	// UntilYourNextTurn is "until your next turn" (CR 611.2b), "you" being
	// Controller; false is "this turn" (CR 514.2).
	UntilYourNextTurn bool

	// Label is the record's label.
	Label string
}

// RedirectDamageThisTurnForEffect registers the redirection. Reports
// whether a record was written: nothing is registered for a destination
// that is not a player in the game or a permanent on the battlefield (the
// effect would do nothing, CR 614.9), a protected permanent that is not
// on the battlefield, or a shape redirectDamageModProblem refuses.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) RedirectDamageThisTurnForEffect(r DamageRedirection) bool {
	m := Mod{
		Kind:               ModRedirectDamage,
		SourceZone:         r.SourceZone,
		Queries:            clonePermanentQueries(r.Queries),
		Player:             r.ProtectPlayer,
		Types:              copyStrings(r.ProtectTypes),
		Recipients:         r.ProtectRecipients,
		CombatOnly:         r.CombatOnly,
		Next:               r.Next,
		Amount:             r.Amount,
		ToSourceController: r.ToSourceController,
		Then:               r.Then.key,
	}
	if r.Source.ID != uuid.Nil {
		m.Objects = []ObjectRef{r.Source}
	} else {
		m.SourceZone = ""
	}
	if r.To != uuid.Nil {
		to, ok := g.redirectRefLocked(r.To)
		if !ok {
			return false
		}
		m.To = []ObjectRef{to}
	}
	if redirectDamageModProblem(m) != "" {
		return false
	}
	label := r.Label
	if label == "" {
		label = "damage dealt to something else instead"
	}
	d := g.UntilEndOfTurnDuration()
	if r.UntilYourNextTurn {
		d = g.UntilYourNextTurnDuration(r.Controller)
	}
	if r.ProtectPermanent != uuid.Nil {
		affected := g.PinnedObjectsLocked(r.ProtectPermanent)
		if len(affected) == 0 {
			return false
		}
		return g.appendScopedEffectLocked(r.EffectSource, affected, ScopeNone, r.Controller, []Mod{m},
			g.PinnedTo(d, r.ProtectPermanent), label, timeNowUnixNano())
	}
	return g.RegisterScopedRuleEffectForEffect(r.EffectSource, ScopeGame, r.Controller, []Mod{m}, d, label)
}

// redirectRefLocked pins a destination as the object it is now: a player
// still in the game (epoch zero), or a permanent on the battlefield at its
// current epoch (CR 400.7).
//
// Caller must hold g.mu.
func (g *Game) redirectRefLocked(id uuid.UUID) (ObjectRef, bool) {
	if p := g.playerByIDLocked(id); p != nil {
		return ObjectRef{ID: id}, !p.Eliminated
	}
	if c := findBattlefieldCard(g, id); c != nil {
		return ObjectRef{ID: id, Epoch: c.ObjectEpoch}, true
	}
	return ObjectRef{}, false
}

// redirectDamageModProblem is registration's and restore's check on a
// ModRedirectDamage mod, and the refusal of its own field on every other
// kind (a newer binary's shape). "" when the mod is sound.
func redirectDamageModProblem(m Mod) string {
	if m.Kind != ModRedirectDamage {
		if m.ToSourceController {
			return fmt.Sprintf("mod %q carries toSourceController, which only redirectDamage reads", m.Kind)
		}
		return ""
	}
	switch {
	case len(m.Objects) > 1:
		return "a redirectDamage effect names more than one source"
	case len(m.Objects) == 1 && m.Objects[0].ID == uuid.Nil:
		return "a redirectDamage effect names an empty source"
	case len(m.Objects) == 0 && m.SourceZone != "":
		return "a redirectDamage effect names a source zone and no source"
	case slices.ContainsFunc(m.Queries, func(q PermanentQuery) bool { return q.Enchanted }):
		return "a redirectDamage recheck cannot ask whether a source is enchanted"
	case m.Amount < 0:
		return fmt.Sprintf("a redirectDamage effect has a negative charge, %d", m.Amount)
	case m.Next && m.Amount > 0:
		return "a redirectDamage effect is both \"the next time\" and a charge"
	case !m.Next && (m.SpentBatch != 0 || m.SpentInstance != 0):
		return "a redirectDamage effect that is not \"the next time\" is marked spent"
	case m.Half:
		return "a redirectDamage effect carries half, which only preventNextFromSource reads"
	case !m.Recipients.Known():
		return fmt.Sprintf("damage recipients %q", m.Recipients)
	case m.Recipients != DamageRecipientsAny && (m.Player != uuid.Nil || len(m.Types) > 0 ||
		m.Recipients == DamageRecipientsPlayerAndTheirPermanents):
		return "a redirectDamage effect's recipients stand alone, with no protected player"
	case len(m.To) > 1:
		return fmt.Sprintf("a redirectDamage effect names %d destinations; at most one", len(m.To))
	case len(m.To) == 1 && m.To[0].ID == uuid.Nil:
		return "a redirectDamage effect names an empty destination"
	case len(m.To) == 1 && m.ToSourceController:
		return "a redirectDamage effect names both a destination and the source's controller"
	case !redirectsAnywhere(m) && m.Then == "":
		return "a redirectDamage effect deals its damage nowhere instead and has no follow-up"
	case m.Then != "" && !KnownEffectBody(m.Then):
		return fmt.Sprintf("redirectDamage names follow-up body %q, which is not registered", m.Then)
	}
	return ""
}

// redirectsAnywhere reports whether a redirection deals its damage to
// something else — false for Eye for an Eye's shape, which only owes its
// follow-up.
func redirectsAnywhere(m Mod) bool { return len(m.To) == 1 || m.ToSourceController }

// redirectDamageAppliesLocked is the kind's applicability, before a CR
// 615.7 division is consulted (scopedReplacementAppliesLocked adds that
// for a charged redirection): damage from the source to what the effect
// protects, which can be dealt to the destination instead now (CR 614.9 —
// a redirection that would do nothing does not apply, so it is not used
// up, CR 609.7b).
//
// Caller must hold g.mu.
func (g *Game) redirectDamageAppliesLocked(e ScopedEffect, m Mod, ev *ReplacementEvent) bool {
	// CR 120.8: nothing to redirect.
	if ev.Kind != RepEventDamage || ev.DamageAmount <= 0 {
		return false
	}
	if m.CombatOnly && !ev.IsCombatDamage {
		return false
	}
	if m.Next && !nextShieldOpenToLocked(g, m, ev) {
		return false
	}
	if len(m.Objects) == 1 && !g.damageFromChosenSourceLocked(m, ev.DamageSource) {
		return false
	}
	// CR 609.7b: the properties are rechecked as the damage would be
	// dealt, against the source as it is then.
	if len(m.Queries) > 0 && !queriesMatchCharacteristic(m.Queries, ev.SourceLKI) {
		return false
	}
	if !g.redirectProtectsLocked(e, m, ev.DamageTarget) {
		return false
	}
	if !redirectsAnywhere(m) {
		return true
	}
	to, ok := g.redirectDestinationLocked(m, ev)
	return ok && g.canRedirectDamageLocked(ev, to)
}

// redirectProtectsLocked reports whether damage to `target` is damage the
// redirection takes: the protected player and the permanents of the named
// types they control, the pinned permanent — and, beside it, the player
// ("this creature and/or you", Glarecaster) — the recipients its
// vocabulary names ("an opponent"), or anything at all when nothing is
// named.
//
// Caller must hold g.mu.
func (g *Game) redirectProtectsLocked(e ScopedEffect, m Mod, target uuid.UUID) bool {
	if m.Recipients != DamageRecipientsAny {
		return g.damageRecipientsMatchLocked(e, m, target)
	}
	if e.Scope != ScopeGame && m.Player != uuid.Nil && target == m.Player {
		return true
	}
	return g.nextFromSourceProtectsLocked(e, m, target)
}

// redirectDestinationLocked is where a redirection deals the damage: its
// To while it is still the object it named (a player still in the game,
// the same permanent, CR 400.7), or the source's controller as the source
// was when the damage would be dealt (its last-known information).
//
// Caller must hold g.mu.
func (g *Game) redirectDestinationLocked(m Mod, ev *ReplacementEvent) (uuid.UUID, bool) {
	if m.ToSourceController {
		if ev.SourceLKI == nil || ev.SourceLKI.Controller == uuid.Nil {
			return uuid.Nil, false
		}
		return ev.SourceLKI.Controller, true
	}
	if len(m.To) != 1 {
		return uuid.Nil, false
	}
	to := m.To[0]
	if p := g.playerByIDLocked(to.ID); p != nil {
		return to.ID, !p.Eliminated
	}
	c := findBattlefieldCard(g, to.ID)
	if c == nil || c.ObjectEpoch != to.Epoch {
		return uuid.Nil, false
	}
	return to.ID, true
}

// damageableRecipientLocked is CR 614.9's test of one end of a
// redirection: a player who has not left the game, or a permanent on the
// battlefield that is still a battle, creature or planeswalker. It reports
// the tail kind damage to it lands through.
//
// Caller must hold g.mu.
func (g *Game) damageableRecipientLocked(id uuid.UUID) (damageTailKind, bool) {
	if id == uuid.Nil {
		return 0, false
	}
	if p := g.playerByIDLocked(id); p != nil {
		return damageTailPlayer, !p.Eliminated
	}
	c := findBattlefieldCard(g, id)
	if c == nil {
		return 0, false
	}
	return damageTailPermanent, c.IsCreature() || c.HasCardType("planeswalker") || c.HasCardType("battle")
}

// canRedirectDamageLocked reports whether the damage event can be dealt to
// `to` instead right now (CR 614.9): both ends are still something damage
// can be dealt to, `to` is not already the recipient, and the event is one
// of the engine's rules damage events (a sandbox mark is not).
//
// Caller must hold g.mu.
func (g *Game) canRedirectDamageLocked(ev *ReplacementEvent, to uuid.UUID) bool {
	if ev == nil || ev.Kind != RepEventDamage || ev.DamageAmount <= 0 || to == ev.DamageTarget {
		return false
	}
	t := ev.damageTail
	if t == nil || (t.kind != damageTailPlayer && t.kind != damageTailPermanent) {
		return false
	}
	// "That damage can't be … dealt instead to another permanent or
	// player" (Lava Burst, Whippoorwill). settleUnpreventableLocked
	// already settles a gathered redirection under it; this catches the
	// one an earlier effect in a CR 616 chain made newly unredirectable.
	if g.damageCantBeRedirectedLocked(ev) {
		return false
	}
	if _, ok := g.damageableRecipientLocked(ev.DamageTarget); !ok {
		return false
	}
	_, ok := g.damageableRecipientLocked(to)
	return ok
}

// redirectDamageEventLocked is the one primitive every redirection uses
// (§9 decision 1): the event's damage is dealt to `to` instead, its
// recipient and its tail's kind rewritten together, so it lands through
// the path its new recipient takes. Everything else about the damage —
// its source, its amount, combat or not, the source's deathtouch,
// lifelink, infect and wither, the commander tally — rides on unchanged
// (CR 120.4b). Reports false, changing nothing, when CR 614.9 says the
// redirection does nothing.
//
// Caller must hold g.mu (write).
func (g *Game) redirectDamageEventLocked(ev *ReplacementEvent, to uuid.UUID) bool {
	if !g.canRedirectDamageLocked(ev, to) {
		return false
	}
	kind, _ := g.damageableRecipientLocked(to)
	ev.DamageTarget = to
	ev.damageTail.kind = kind
	return true
}

// RedirectDamageEventForEffect is the primitive's catalog face: a static
// redirection's Replace ("All damage that would be dealt to you is dealt
// to enchanted creature instead", Pariah) calls it with the event it was
// handed. The ReplacementEffect must declare RedirectsDamage (the census
// test holds it). Reports whether the damage was redirected.
//
// Caller must hold g.mu (write) — every caller is a Replace.
func (g *Game) RedirectDamageEventForEffect(ev *ReplacementEvent, to uuid.UUID) bool {
	return g.redirectDamageEventLocked(ev, to)
}

// CanRedirectDamageForEffect is canRedirectDamageLocked for a static
// redirection's AppliesTo: a redirection that would do nothing (CR 614.9)
// does not apply, so it is never offered in a CR 616 ordering prompt.
//
// Caller must hold g.mu.
func (g *Game) CanRedirectDamageForEffect(ev *ReplacementEvent, to uuid.UUID) bool {
	return g.canRedirectDamageLocked(ev, to)
}

// redirectedDamage is the part of a damage event a charged redirection
// split off: `amount` of it, dealt to `to` instead, with the effects
// already applied to the event (CR 614.5). Rides damageTail.redirected
// until the event reaches its terminal outcome. Transient, never captured.
type redirectedDamage struct {
	to      uuid.UUID
	amount  int
	applied []ReplacementEffectID
}

// redirectDamagePartLocked redirects `n` of the event's damage and leaves
// the rest where it was: the event is reduced by n, and the n rides its
// tail to be dealt to `to` as the event ends (dealRedirectedDamageLocked).
// n covering the whole event is the primitive itself. Reports false,
// changing nothing, when CR 614.9 says the redirection does nothing.
//
// Caller must hold g.mu (write).
func (g *Game) redirectDamagePartLocked(ev *ReplacementEvent, to uuid.UUID, n int) bool {
	if n <= 0 {
		return false
	}
	if n >= ev.DamageAmount {
		return g.redirectDamageEventLocked(ev, to)
	}
	if !g.canRedirectDamageLocked(ev, to) {
		return false
	}
	var applied []ReplacementEffectID
	for id, done := range g.replacementsAppliedThisEvent[ev.ID] {
		if done {
			applied = append(applied, id)
		}
	}
	slices.Sort(applied)
	ev.DamageAmount -= n
	t := ev.damageTail
	t.redirected = append(slices.Clip(t.redirected), redirectedDamage{to: to, amount: n, applied: applied})
	return true
}

// dealRedirectedDamageLocked deals the parts charged redirections split
// off a settled event, each as its own damage event of the same instance,
// from the same source, through the CR 614 window again — the new
// recipient's shields and protection apply (CR 616.2) and nothing already
// applied to the event applies twice (CR 614.5). Called from
// runDamageTailLocked, on every terminal outcome of the event, before the
// caller's continuation and before the instance ends.
//
// Caller must hold g.mu (write).
func (g *Game) dealRedirectedDamageLocked(ev *ReplacementEvent) {
	t := ev.damageTail
	parts := t.redirected
	t.redirected = nil
	for _, p := range parts {
		kind, ok := g.damageableRecipientLocked(p.to)
		if !ok {
			continue
		}
		tail := *t
		tail.kind = kind
		tail.then = nil
		tail.endsInstance = false
		tail.released = true
		tail.redirected = nil
		part := &ReplacementEvent{
			Kind:           RepEventDamage,
			Source:         ev.Source,
			DamageSource:   ev.DamageSource,
			DamageTarget:   p.to,
			DamageAmount:   p.amount,
			IsCombatDamage: ev.IsCombatDamage,
			DamageInstance: ev.DamageInstance,
			SourceLKI:      ev.SourceLKI,
			damageTail:     &tail,
		}
		part.ID = ReplacementEventID(g.nextReplacementEventID.Add(1))
		if g.replacementsAppliedThisEvent == nil {
			g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
		}
		applied := make(map[ReplacementEffectID]bool, len(p.applied))
		for _, id := range p.applied {
			applied[id] = true
		}
		g.replacementsAppliedThisEvent[part.ID] = applied
		if _, err := g.damageThroughReplacementsLocked(part); err != nil {
			// The new recipient is gone: the part is simply not dealt.
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: "redirected damage: " + err.Error()})
		}
	}
}

// applyRedirectDamageLocked is the kind's Replace: the damage — all of
// it, or as much of it as the charge (and the share, when the protected
// player divided it) allows — is dealt to the destination instead, and a
// "next time" redirection is marked spent with this instance. Eye for an
// Eye's shape leaves the event as it is and owes its follow-up with the
// amount.
//
// Caller must hold g.mu (write).
func (g *Game) applyRedirectDamageLocked(e ScopedEffect, mod int, m Mod, ev *ReplacementEvent) {
	damage := ev.DamageAmount
	if !redirectsAnywhere(m) {
		g.markRedirectSpentLocked(e, mod, m, ev)
		g.queuePreventionFollowUpLocked(e, m, ev, 0, damage)
		return
	}
	to, ok := g.redirectDestinationLocked(m, ev)
	if !ok || !g.canRedirectDamageLocked(ev, to) {
		// CR 614.9: the effect does nothing, and is not used up
		// (CR 609.7b).
		return
	}
	n := damage
	if m.Amount > 0 {
		n = g.spendChargeLocked(e, mod, m, ev)
		if n <= 0 {
			return
		}
	}
	g.markRedirectSpentLocked(e, mod, m, ev)
	g.redirectDamagePartLocked(ev, to, n)
	g.queuePreventionFollowUpLocked(e, m, ev, 0, n)
}

// markRedirectSpentLocked spends a "next time" redirection with ev's
// instance (CR 615.8's reading, ADR 0108's one spend rule).
//
// Caller must hold g.mu (write).
func (g *Game) markRedirectSpentLocked(e ScopedEffect, mod int, m Mod, ev *ReplacementEvent) {
	if m.Next && m.SpentBatch == 0 && m.SpentInstance == 0 {
		g.markNextShieldSpentLocked(e.Seq, mod, g.currentEventBatchLocked(), ev.DamageInstance)
	}
}

// DamageRedirectionLines is the game banner's line for each live
// redirection, oldest first (ADR 0108 "Shared machinery", the client):
// "Damage to Alice from Goblin Guide is dealt to Beacon of Destiny
// instead, the next time — Beacon of Destiny". Empty on nearly every turn.
//
// Caller must hold g.mu (read or write).
func (g *Game) DamageRedirectionLines() []string {
	var out []string
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModRedirectDamage {
				continue
			}
			out = append(out, g.damageRedirectionLineLocked(e, m))
		}
	}
	return out
}

// damageRedirectionLineLocked writes one record's banner line.
func (g *Game) damageRedirectionLineLocked(e *ScopedEffect, m Mod) string {
	name := func(id uuid.UUID) string {
		if p := g.playerByIDLocked(id); p != nil {
			return p.Name
		}
		if c, ok := g.LookupCardForEffect(id); ok && c.Name != "" && !c.FaceDown {
			return c.Name
		}
		return "a permanent"
	}
	var b strings.Builder
	if m.CombatOnly {
		b.WriteString("Combat damage")
	} else {
		b.WriteString("Damage")
	}
	switch {
	case e.Scope == ScopeNone && len(e.Affected) == 1:
		b.WriteString(" to " + name(e.Affected[0].ID))
	case m.Player != uuid.Nil && len(m.Types) > 0:
		b.WriteString(" to " + name(m.Player) + " and their permanents")
	case m.Player != uuid.Nil:
		b.WriteString(" to " + name(m.Player))
	}
	if len(m.Objects) == 1 {
		b.WriteString(" from " + name(m.Objects[0].ID))
	}
	switch {
	case len(m.To) == 1:
		b.WriteString(" is dealt to " + name(m.To[0].ID) + " instead")
	case m.ToSourceController:
		b.WriteString(" is dealt to its source's controller instead")
	default:
		b.WriteString(" is answered")
	}
	switch {
	case m.Next:
		b.WriteString(", the next time")
	case m.Amount > 0:
		b.WriteString(fmt.Sprintf(", the next %d", m.Amount))
	}
	if n := scopedEffectDisplayName(e); n != "" {
		b.WriteString(" — " + n)
	}
	return b.String()
}
