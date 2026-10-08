package effects

import (
	"fmt"
	"slices"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// purpose.go — ADR 0126 §6: the registration guard for a declared
// Purpose. The type and what each field means are game/purpose.go.
//
// A card file declares a purpose in one of five slots:
//
//	Spec.Purpose                              the spell, or a permanent's enters effect
//	Spec.Modes.Options[i].Purpose             one bullet of a modal spell
//	Spec.AlternativeCosts[i].Purpose          the spell cast for that cost (overload)
//	Spec.Activated[i].Purpose                 an activated ability
//	Spec.Triggered[i].Purpose                 a triggered ability
//
// plus the modes of an activated or triggered ability. Like
// Completeness it is written by hand from the printed text, and the
// engine never reads it: it rides the wire for the bot (ADR 0033 §3).
//
// The guard refuses what nothing would read or what cannot be true,
// each at boot rather than as a bot that quietly misprices a card:
//
//   - a negative amount;
//   - ControllerLosesLife anywhere but an any-player activated row,
//     where the activator and the controller differ (ADR 0106);
//   - DeathPayoff anywhere but a triggered row;
//   - a DiscardPayoff anywhere but a triggered row, or one that names
//     no cards, names them two ways, or pays nothing (ADR 0126's
//     amendment of 2026-10-06);
//   - a Pump or PreventCombatDamageToSelf anywhere but a triggered or
//     activated row, or a Pump that gives nothing (ADR 0130's
//     amendment of 2026-10-07);
//   - a Sweep with an unknown class or verb, an amount on a verb that
//     has none, or none on a verb that needs one.

// purposeSlot says where a Purpose was declared, for the guard's rules.
type purposeSlot int

const (
	purposeOnCard purposeSlot = iota
	purposeOnMode
	purposeOnAltCost
	purposeOnActivated
	purposeOnAnyPlayerActivated
	purposeOnTriggered
)

// checkPurpose panics when p is not a legal declaration for its slot.
// `where` names the slot for the message ("mode 2", "ability 0").
func checkPurpose(name, where string, slot purposeSlot, p game.Purpose) {
	if p.IsZero() {
		return
	}
	fail := func(why string) {
		panic(fmt.Sprintf("effects.Register: %q %s declares a Purpose that %s (ADR 0126 §6)", name, where, why))
	}
	if p.Draws < 0 || p.ControllerLosesLife < 0 || p.Discards < 0 || p.Lands < 0 ||
		p.Tutors < 0 || p.SelfMillTutor < 0 || p.Tokens < 0 || p.Energy < 0 || p.Sweep.Amount < 0 ||
		p.ExtraCombat < 0 || p.DamageToCreature < 0 || p.DamageEachOpponent < 0 || p.LifeGain < 0 {
		fail("has a negative amount")
	}
	if p.ControllerLosesLife != 0 && slot != purposeOnAnyPlayerActivated {
		fail("names ControllerLosesLife off an any-player activated row — the activator and the controller are one player there")
	}
	if p.DeathPayoff && slot != purposeOnTriggered {
		fail("sets DeathPayoff off a triggered row")
	}
	if p.DiscardPayoff != nil && slot != purposeOnTriggered {
		fail("sets DiscardPayoff off a triggered row")
	}
	onSource := slot == purposeOnTriggered || slot == purposeOnActivated || slot == purposeOnAnyPlayerActivated
	if p.Pump != nil && !onSource {
		fail("sets Pump off a triggered or activated row — a pump is what an ability gives its own source")
	}
	if p.PreventCombatDamageToSelf && !onSource {
		fail("sets PreventCombatDamageToSelf off a triggered or activated row")
	}
	checkPump(p.Pump, fail)
	checkDiscardPayoff(p.DiscardPayoff, fail)
	checkSweep(p.Sweep, fail)
}

// checkPump is checkPurpose's half for a declared Pump (ADR 0130's
// amendment of 2026-10-07): it gives something, and each keyword is one
// lowercase word, as the wire's `abilities` spells it.
func checkPump(pm *game.Pump, fail func(string)) {
	if pm == nil {
		return
	}
	if pm.Power == 0 && pm.Toughness == 0 && len(pm.Keywords) == 0 {
		fail("is a pump that gives nothing — set Power, Toughness or Keywords")
	}
	for _, k := range pm.Keywords {
		if k == "" || k != strings.ToLower(k) || strings.ContainsAny(k, "\t—") {
			fail(fmt.Sprintf("names a pump keyword %q that is not lowercase", k))
		}
	}
}

// checkDiscardPayoff is checkPurpose's half for a declared
// DiscardPayoff: it names the cards it pays on one way (Any, or a
// non-empty list of lowercase types), and it pays something.
func checkDiscardPayoff(d *game.DiscardPayoff, fail func(string)) {
	if d == nil {
		return
	}
	switch {
	case d.Any && len(d.Types) > 0:
		fail("pays on discards both as Any and by Types — pick one")
	case !d.Any && len(d.Types) == 0:
		fail("pays on no discarded card — set Any, or name the Types")
	}
	for _, t := range d.Types {
		if t == "" || t != strings.ToLower(t) || strings.ContainsAny(t, " \t—") {
			fail(fmt.Sprintf("names a discard payoff type %q that is not one lowercase word", t))
		}
	}
	if d.Tokens < 0 || d.Counters < 0 || d.DamageEachOpponent < 0 {
		fail("has a negative amount")
	}
	if d.Tokens == 0 && d.Counters == 0 && d.DamageEachOpponent == 0 {
		fail("is a discard payoff that pays nothing — set Tokens, Counters or DamageEachOpponent")
	}
}

// checkSweep is checkPurpose's half for a declared Sweep.
func checkSweep(s game.Sweep, fail func(string)) {
	if s.IsZero() {
		return
	}
	if !slices.Contains(game.SweepMatches, s.Matches) {
		fail(fmt.Sprintf("sweeps an unknown class %q", s.Matches))
	}
	if !slices.Contains(game.SweepHows, s.How) {
		fail(fmt.Sprintf("sweeps in an unknown way %q", s.How))
	}
	sized := s.How == game.SweepDamage || s.How == game.SweepMinus
	switch {
	case sized && s.Amount == 0 && !s.AmountIsX:
		fail(fmt.Sprintf("sweeps by %s with no amount — set Amount, or AmountIsX for an X", s.How))
	case sized && s.Amount != 0 && s.AmountIsX:
		fail("sets both Amount and AmountIsX")
	case !sized && (s.Amount != 0 || s.AmountIsX):
		fail(fmt.Sprintf("gives an amount to a %s sweep, which has none", s.How))
	}
}

// checkModePurposes checks every bullet of a mode clause. A bullet is
// a mode slot whatever owns it: a spell, an activated or a triggered
// ability.
func checkModePurposes(name, where string, m *game.ModeSpec) {
	if m == nil {
		return
	}
	for i, o := range m.Options {
		checkPurpose(name, fmt.Sprintf("%s mode %d", where, i), purposeOnMode, o.Purpose)
	}
}

// checkSpecPurposes is Register's purpose guard over a whole Spec.
func checkSpecPurposes(spec Spec) {
	name := spec.Name
	checkPurpose(name, "card", purposeOnCard, spec.Purpose)
	checkModePurposes(name, "spell", spec.Modes)
	for _, a := range spec.AlternativeCosts {
		checkPurpose(name, fmt.Sprintf("alternative cost %q", a.Key), purposeOnAltCost, a.Purpose)
	}
	for i, a := range spec.Activated {
		checkActivatedPurpose(name, fmt.Sprintf("ability %d", i), a)
	}
	for i, t := range spec.Triggered {
		where := fmt.Sprintf("triggered ability %d", i)
		checkPurpose(name, where, purposeOnTriggered, t.Purpose)
		checkModePurposes(name, where, t.Modes)
	}
}

// checkActivatedPurpose checks one activated row's purpose and its
// modes'. Shared with the ability-grant guard, because a granted row
// is declared exactly as a card's own is.
func checkActivatedPurpose(name, where string, a ActivatedAbility) {
	slot := purposeOnActivated
	if a.AnyPlayer {
		slot = purposeOnAnyPlayerActivated
	}
	checkPurpose(name, where, slot, a.Purpose)
	checkModePurposes(name, where, a.Modes)
}

// ModeWithPurpose is a mode bullet with its declared purpose: the
// modal form of Spec.Purpose (ADR 0126 §6). Wrap the bullet the mode
// helpers build:
//
//	Modes: ChooseN("Choose one or more", 1, 4,
//		ModeWithPurpose(Mode("Exile all creatures."), game.Purpose{Sweep: ...}),
//		...
func ModeWithPurpose(o game.ModeOption, p game.Purpose) game.ModeOption {
	o.Purpose = p
	return o
}

// CostWithPurpose is an alternative cost with what the spell does when
// cast for it (ADR 0126 §6): overload's "each" turns a removal spell
// into a sweep, so the purpose rides on the overload offer rather than
// on the card.
//
//	AlternativeCosts: []game.AlternativeCost{
//		CostWithPurpose(Overload("{6}{U}"), game.Purpose{Sweep: ...}),
//	},
func CostWithPurpose(ac game.AlternativeCost, p game.Purpose) game.AlternativeCost {
	ac.Purpose = p
	return ac
}

// TriggerWithPurpose is a triggered row with its declared purpose
// (ADR 0126 §6): DeathPayoff on "whenever a creature you control
// dies", a Sweep on a chapter that destroys all creatures. Wrap the row
// the trigger helpers build:
//
//	WheneverACreatureYouControlDies(...) becomes
//	TriggerWithPurpose(WheneverACreatureYouControlDies(...), game.Purpose{DeathPayoff: true})
//
// A discard payoff is declared the same way:
//
//	TriggerWithPurpose(On(game.EventDiscardCard, ...), game.Purpose{
//		DiscardPayoff: &game.DiscardPayoff{Any: true, Counters: 1}})
func TriggerWithPurpose(t game.TriggeredAbility, p game.Purpose) game.TriggeredAbility {
	t.Purpose = p
	return t
}
