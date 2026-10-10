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
//   - more LandsUntapped than Lands (ADR 0136 §2);
//   - ControllerLosesLife anywhere but an any-player activated row,
//     where the activator and the controller differ (ADR 0106);
//   - DeathPayoff anywhere but a triggered row;
//   - a DiscardPayoff anywhere but a triggered row, or one that names
//     no cards, names them two ways, or pays nothing (ADR 0126's
//     amendment of 2026-10-06);
//   - a Pump or PreventCombatDamageToSelf anywhere but a triggered or
//     activated row, or a Pump that gives nothing (ADR 0130's
//     amendment of 2026-10-07);
//   - AwakenLand anywhere but an alternative cost (ADR 0135 §3);
//   - ExtraLandDrops anywhere but the card, or on a card whose
//     AdditionalLandPlays says a different number (#2678);
//   - a Sweep with an unknown class or verb, an amount on a verb that
//     has none, or none on a verb that needs one;
//   - a target entry (Purpose.Targets, ADR 0126's amendment of
//     2026-10-08) that names no clause of its statement, names one
//     twice, says nothing, or says what the clause's pick cannot be
//     given (checkTargetPurposes).
//   - Answers (ADR 0142) off an activated row, or a declaration
//     answers.go refuses (checkActivatedAnswers, checkManaAnswers).

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
		p.ExtraCombat < 0 || p.DamageToCreature < 0 || p.DamageEachOpponent < 0 || p.LifeGain < 0 ||
		p.AwakenLand < 0 || p.ExtraLandDrops < 0 || p.LandsUntapped < 0 {
		fail("has a negative amount")
	}
	if p.LandsUntapped > p.Lands {
		fail("says more lands enter untapped (LandsUntapped) than it puts onto the battlefield (Lands) (ADR 0136 §2)")
	}
	if p.ExtraLandDrops != 0 && slot != purposeOnCard {
		fail("sets ExtraLandDrops off the card — an additional land drop is the card's own static or its spell's effect (#2678)")
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
	if p.Answers != 0 && !onActivated(slot) {
		fail("declares Answers off an activated row — nothing reads it there yet (ADR 0142 §2)")
	}
	checkAnswerBits(p.Answers, fail)
	if p.AwakenLand != 0 && slot != purposeOnAltCost {
		fail("sets AwakenLand off an alternative cost — it is what an awaken offer adds (ADR 0135 §3)")
	}
	checkPump(p.Pump, fail)
	checkDiscardPayoff(p.DiscardPayoff, fail)
	checkSweep(p.Sweep, fail)
}

// onActivated reports whether a slot is an activated row, the one slot
// Answers may be declared on (ADR 0142 §2).
func onActivated(slot purposeSlot) bool {
	return slot == purposeOnActivated || slot == purposeOnAnyPlayerActivated
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
		mw := fmt.Sprintf("%s mode %d", where, i)
		checkPurpose(name, mw, purposeOnMode, o.Purpose)
		checkTargetPurposes(name, mw, o.Purpose, o.Targets, true)
	}
}

// checkSpecPurposes is Register's purpose guard over a whole Spec.
func checkSpecPurposes(spec Spec) {
	name := spec.Name
	checkPurpose(name, "card", purposeOnCard, spec.Purpose)
	if n := spec.Purpose.ExtraLandDrops; n != 0 && spec.AdditionalLandPlays != 0 && n != spec.AdditionalLandPlays {
		panic(fmt.Sprintf("effects.Register: %q declares ExtraLandDrops %d but AdditionalLandPlays %d: the purpose says what the static does (#2678)",
			name, n, spec.AdditionalLandPlays))
	}
	checkTargetPurposes(name, "card", spec.Purpose, spec.Targets, true)
	checkModePurposes(name, "spell", spec.Modes)
	for _, a := range spec.AlternativeCosts {
		where := fmt.Sprintf("alternative cost %q", a.Key)
		checkPurpose(name, where, purposeOnAltCost, a.Purpose)
		// The statement a cast for this cost announces: its own clause
		// list when it replaces the card's, none when it clears it
		// (overload), else the card's.
		stmt := spec.Targets
		switch {
		case a.Targets != nil:
			stmt = a.Targets
		case a.ClearsTargets:
			stmt = nil
		}
		checkTargetPurposes(name, where, a.Purpose, stmt, true)
	}
	for i, a := range spec.Activated {
		checkActivatedPurpose(name, fmt.Sprintf("ability %d", i), a)
	}
	for i, t := range spec.Triggered {
		checkTriggeredPurpose(name, fmt.Sprintf("triggered ability %d", i), t)
	}
	checkManaAbilitiesAnswers(name, "card", spec.ManaAbilities)
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
	checkActivatedAnswers(name, where, activatedShapeOf(a))
	checkTargetPurposes(name, where, a.Purpose, a.Targets, true)
	checkModePurposes(name, where, a.Modes)
}

// checkTriggeredPurpose checks one triggered row's purpose and its
// modes'. Shared with the ability-grant guard. A row whose clause list
// is built at trigger time (TargetsFrom) has no statement to check a
// target entry against, so it may not declare one.
func checkTriggeredPurpose(name, where string, t game.TriggeredAbility) {
	checkPurpose(name, where, purposeOnTriggered, t.Purpose)
	checkTargetPurposes(name, where, t.Purpose, t.Targets, t.Targets != nil || t.TargetsFrom == nil)
	checkModePurposes(name, where, t.Modes)
}

// checkTargetPurposes is the guard for Purpose.Targets (ADR 0126's
// amendment of 2026-10-08): each entry names a target clause of the
// statement it is declared on, once, and says something that clause's
// pick can be given. `stmt` is that statement's clause list; `known`
// is false when the list is only built at run time.
//
// It refuses:
//
//   - an empty list (declare nil instead);
//   - a slot that is not a clause of the statement, or one named twice;
//   - an entry that says nothing, or has a negative amount;
//   - a player amount (draws, discards, tokens, life gained or lost)
//     on a clause that cannot target a player;
//   - damage on a clause that can target neither a player nor a
//     permanent;
//   - a return to the target's controller (#2679) on a clause that
//     cannot target a permanent, a negative one, creature tokens with
//     no printed toughness or a size with no tokens, or more untapped
//     lands than lands. The clause's card predicate is a closure, so "target
//     artifact" is a permanent clause to this check; the card's text
//     is the rest of the review.
func checkTargetPurposes(name, where string, p game.Purpose, stmt *game.TargetSpec, known bool) {
	if p.Targets == nil {
		return
	}
	fail := func(why string) {
		panic(fmt.Sprintf("effects.Register: %q %s declares a target purpose that %s (ADR 0126, amendment of 2026-10-08)", name, where, why))
	}
	entries := p.Targets.List()
	if len(entries) == 0 {
		fail("is an empty list — leave Targets nil")
	}
	if !known {
		fail("is declared on a row whose target clauses are built when it triggers (TargetsFrom), so no slot can be checked")
	}
	seen := map[int]bool{}
	for _, e := range entries {
		if e.Slot < 0 || e.Slot >= stmt.ClauseCount() {
			fail(fmt.Sprintf("names slot %d, which is not one of the statement's %d target clause(s)", e.Slot, stmt.ClauseCount()))
		}
		if seen[e.Slot] {
			fail(fmt.Sprintf("names slot %d twice", e.Slot))
		}
		seen[e.Slot] = true
		if e.Draws < 0 || e.Discards < 0 || e.Tokens < 0 || e.LifeGain < 0 || e.LifeLoss < 0 || e.Damage < 0 {
			fail(fmt.Sprintf("has a negative amount on slot %d", e.Slot))
		}
		if e.IsZero() {
			fail(fmt.Sprintf("says nothing about slot %d", e.Slot))
		}
		clause := stmt.Clause(e.Slot)
		if e.HasPlayerAmount() && !clause.Players {
			fail(fmt.Sprintf("gives a player amount to slot %d, whose clause %q cannot target a player", e.Slot, clause.Label))
		}
		if e.DamageIsX && e.Damage != 0 {
			fail(fmt.Sprintf("sets both Damage and DamageIsX on slot %d", e.Slot))
		}
		if (e.Damage != 0 || e.DamageIsX) && !clause.Players && !slices.Contains(clause.Zones, game.ZoneBattlefield) {
			fail(fmt.Sprintf("deals damage to slot %d, whose clause %q can target neither a player nor a permanent", e.Slot, clause.Label))
		}
		if r := e.Returns; !r.IsZero() {
			if !slices.Contains(clause.Zones, game.ZoneBattlefield) {
				fail(fmt.Sprintf("returns something to the controller of slot %d, whose clause %q cannot target a permanent", e.Slot, clause.Label))
			}
			if r.CreatureTokens < 0 || r.TokenPower < 0 || r.TokenToughness < 0 || r.Lands < 0 || r.LandsUntapped < 0 {
				fail(fmt.Sprintf("has a negative return on slot %d", e.Slot))
			}
			if (r.CreatureTokens == 0) != (r.TokenPower == 0 && r.TokenToughness == 0) || (r.CreatureTokens > 0 && r.TokenToughness == 0) {
				fail(fmt.Sprintf("returns creature tokens on slot %d without a printed toughness, or a token size with no tokens", e.Slot))
			}
			if r.LandsUntapped > r.Lands {
				fail(fmt.Sprintf("returns more untapped lands than lands on slot %d", e.Slot))
			}
		}
	}
}

// ForTargets is a Purpose that declares only what happens to the
// statement's targets (ADR 0126's amendment of 2026-10-08), one entry
// per target clause:
//
//	Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2}),
func ForTargets(entries ...game.TargetPurpose) game.Purpose {
	return game.Purpose{Targets: game.ForTargets(entries...)}
}

// RemovalReturning is the target entry for a removal whose target's
// controller is given something back (#2679): Rapid Hybridization is
// ForTargets(RemovalReturning(0, game.TargetReturn{CreatureTokens: 1,
// TokenPower: 3, TokenToughness: 3})).
func RemovalReturning(slot int, r game.TargetReturn) game.TargetPurpose {
	return game.TargetPurpose{Slot: slot, Returns: r}
}

// DamageToTarget is the target entry for "deals n damage to" the pick
// of clause `slot`: Lightning Bolt is ForTargets(DamageToTarget(0, 3)).
func DamageToTarget(slot, n int) game.TargetPurpose {
	return game.TargetPurpose{Slot: slot, Damage: n}
}

// DamageXToTarget is the target entry for "deals X damage to" the pick
// of clause `slot`, where X is the announced X (#1944): Chandra,
// Awakened Inferno's −X is ForTargets(DamageXToTarget(0)).
func DamageXToTarget(slot int) game.TargetPurpose {
	return game.TargetPurpose{Slot: slot, DamageIsX: true}
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
