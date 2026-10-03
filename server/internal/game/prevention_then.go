package game

import (
	"fmt"

	"github.com/google/uuid"
)

// prevention_then.go — ADR 0108 §8 (#1906): a prevention STATIC with an
// additional effect, and the apply loop's measure of what a prevention
// effect prevented.
//
// THE RULES.
//
//   - CR 615.5: "Some prevention effects also include an additional
//     effect, which may refer to the amount of damage that was prevented.
//     The prevention takes place at the time the original event would
//     have happened; the rest of the effect takes place immediately
//     afterward."
//   - CR 615.12: "If unpreventable damage would be dealt, any applicable
//     prevention effects are still applied to it. Those effects won't
//     prevent any damage, but any additional effects they have will take
//     place." Nine Lives still gets its incarnation counter, and a
//     Phantom still loses its +1/+1 counter.
//   - The rulings fix the unit, by the printed subject. Phantom Centaur,
//     "if damage would be dealt to this creature": dealt damage by several
//     sources at once, "all of the damage is prevented and only one +1/+1
//     counter is removed". Nine Lives, "if a source would deal damage to
//     you": "if more than one source deals damage to you at once, prevent
//     the damage from each of them and put that many incarnation counters".
//
// THE SHAPE (§8 decisions 1–5).
//
//   - The card declares ReplacementEffect.Then (a registered body) and
//     ThenPer (PreventionUnit): one application per RECIPIENT or per
//     SOURCE, within one damage instance (damage_instance.go).
//   - The Replace itself only prevents. The apply loop measures the event
//     before and after it (runReplaceLocked): what Replace took off the
//     amount — all of it when it cancelled the event — is what was
//     prevented. TestPreventionStaticsOnlyPrevent (cards/effects) fails a
//     Prevention static whose Replace does anything but change the event,
//     so the measure is the whole truth.
//   - The follow-up joins the one queue every scoped shield's does
//     (PreventionFollowUp), keyed on the object whose static it is, the
//     replacement's slot and the unit, and runs once per key and instance
//     as the instance ends (flushPreventionFollowUpsForInstanceLocked), or
//     at ADR 0107's flush points behind it.
//   - Under CR 615.12 the settle step (settleUnpreventableLocked) owes it
//     with nothing prevented and the whole event as "that damage"
//     (preventionAppliedToUnpreventableLocked).
//
// "THAT MANY" IS NOT "PREVENTED THIS WAY". A follow-up carries both: the
// damage the effect was applied to (the handed event's Amount) and what it
// prevented (the params' Amount). They are the same whenever the static
// prevented the damage, and differ under CR 615.12: Polukranos's ruling
// removes "that many +1/+1 counters" from unpreventable damage, and
// Phyrexian Hydra's puts on no -1/-1 counter "for each 1 damage prevented
// this way". The card's body reads the one its text names.

// PreventionUnit is ReplacementEffect.ThenPer: what one application of a
// prevention static's additional effect is per, within one damage
// instance.
type PreventionUnit string

const (
	// ThenPerRecipient is "if damage would be dealt to X": one
	// application per recipient (the Phantoms, Oathsworn Knight).
	ThenPerRecipient PreventionUnit = "recipient"
	// ThenPerSource is "if a source would deal damage to X": one
	// application per damage source (Nine Lives, Ironscale Hydra).
	ThenPerSource PreventionUnit = "source"
)

// Valid reports whether u is one of the two units.
func (u PreventionUnit) Valid() bool { return u == ThenPerRecipient || u == ThenPerSource }

// runReplaceLocked fires one gathered replacement's Replace — every
// apply-loop path that applies an effect goes through here — and, for a
// prevention static with an additional effect, owes that effect with what
// the Replace prevented (CR 615.5). A Replace error is the effect's own
// failure, logged; the event goes on.
//
// Caller must hold g.mu (write).
func (g *Game) runReplaceLocked(ev *ReplacementEvent, a activeReplacement) {
	if a.effect.Replace == nil {
		return
	}
	measured := g.preventionStaticLocked(ev, a)
	before := ev.DamageAmount
	if err := a.effect.Replace(ev, g, a.source); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
	}
	if !measured {
		return
	}
	after := ev.DamageAmount
	if ev.Canceled {
		after = 0
	}
	g.oweStaticFollowUpLocked(ev, a, max(before-after, 0), before)
}

// preventionStaticLocked reports whether `a` is a prevention static with
// an additional effect, applied to a damage event that still has damage
// in it: the case whose Replace the apply loop measures.
func (g *Game) preventionStaticLocked(ev *ReplacementEvent, a activeReplacement) bool {
	return ev != nil && ev.Kind == RepEventDamage && !ev.Canceled &&
		a.effect.Prevention && a.effect.Then.key != "" && a.source != nil
}

// oweStaticFollowUpLocked owes a prevention static's additional effect
// for one application to `ev`: `prevented` of `damage`. The entry is keyed
// on the static's object, its slot and the unit its ThenPer names
// (PreventionFollowUp.owes), so three blockers' damage to a Phantom is one
// application and three sources' damage to Nine Lives is three.
//
// Caller must hold g.mu (write).
func (g *Game) oweStaticFollowUpLocked(ev *ReplacementEvent, a activeReplacement, prevented, damage int) {
	if a.source == nil || a.effect.Then.key == "" || ev == nil {
		return
	}
	f := g.newPreventionFollowUpLocked(ev, a.effect.Then.key, prevented, damage)
	f.Static = true
	f.Source = a.source.InstanceID
	f.Epoch = a.source.ObjectEpoch
	f.Slot = a.identity.slot
	f.Controller = a.source.Controller
	f.Label = a.effect.Label
	f.Unit = preventionUnitKey(a.effect.ThenPer, ev)
	g.owePreventionFollowUpLocked(f)
}

// followUpModProblem is registration's and restore's check on Mod.To: at
// most one object, a real one, and only on a prevention shield with a
// follow-up to deal its damage — or on a redirection, whose destination
// it is (ADR 0108 §9; redirectDamageModProblem checks that kind). "" when
// it is sound.
func followUpModProblem(m Mod) string {
	if len(m.To) == 0 || m.Kind == ModRedirectDamage {
		return ""
	}
	if len(m.To) > 1 {
		return fmt.Sprintf("a %s mod names %d follow-up recipients; at most one", m.Kind, len(m.To))
	}
	if m.To[0].ID == uuid.Nil {
		return fmt.Sprintf("a %s mod names an empty follow-up recipient", m.Kind)
	}
	if !scopedKindPrevents(m.Kind) || m.Then == "" {
		return fmt.Sprintf("mod %q carries a follow-up recipient (to) and no prevention follow-up to read it", m.Kind)
	}
	return ""
}

// preventionUnitKey is the unit an application is per: the recipient or
// the damage source. An undeclared unit (refused by effects.Register) is
// read as the recipient.
func preventionUnitKey(per PreventionUnit, ev *ReplacementEvent) uuid.UUID {
	if per == ThenPerSource {
		return ev.DamageSource
	}
	return ev.DamageTarget
}
