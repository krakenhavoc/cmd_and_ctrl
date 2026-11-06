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
//   - a LandfallPayoff anywhere but a triggered row, or one that pays
//     nothing or a negative amount (#3072);
//   - a LifeGainPayoff anywhere but a triggered row, one that pays
//     nothing or a negative amount, or one that gains life (#4979);
//   - a Pump or PreventCombatDamageToSelf anywhere but a triggered or
//     activated row, or a Pump that gives nothing (ADR 0130's
//     amendment of 2026-10-07);
//   - AwakenLand anywhere but an alternative cost (ADR 0135 §3);
//   - Aura anywhere but the card, or a value other than benefit or
//     harm (#3309);
//   - ExtraLandDrops anywhere but the card, or on a card whose
//     AdditionalLandPlays says a different number (#2678);
//   - LandfallExtraTriggers anywhere but the card, or on a card with
//     no TriggerDoubler (#3146);
//   - ExtraLoyaltyActivations anywhere but the card or an activated
//     row of the controller's own (#4191);
//   - SourceCounters anywhere but an activated row of the controller's
//     own (#5083);
//   - ResolutionMana anywhere but an activated row of the controller's
//     own, or one with no resolution or no fixed mana (#4980);
//   - a Sweep with an unknown class or verb, an amount on a verb that
//     has none, or none on a verb that needs one;
//   - a target entry (Purpose.Targets, ADR 0126's amendment of
//     2026-10-08) that names no clause of its statement, names one
//     twice, says nothing, or says what the clause's pick cannot be
//     given (checkTargetPurposes).
//   - Answers (ADR 0142) anywhere but an activated row, the card, a
//     spell's mode or an alternative cost (#2877, ADR 0142's amendment
//     of 2026-10-30), other than AnswerRemove on a triggered one (an
//     edict, #3882); inside Kicked or FaceUp; or a declaration
//     answers.go refuses (checkActivatedAnswers, checkManaAnswers,
//     checkSpellAnswers).
//   - Kicked or FaceUp (ADR 0126's amendment of 2026-10-15, #3762) off
//     the card, on a card with no single kicker / no face-down cast
//     offer, or whose inner purpose is empty, nests another, or
//     declares what only the card's own slot may (checkConditionalPurposes).
//   - a negative rider, or one on a triggered row (checkRiders, #4978).

// purposeSlot says where a Purpose was declared, for the guard's rules.
type purposeSlot int

const (
	purposeOnCard purposeSlot = iota
	// purposeOnMode is a bullet of an activated or triggered ability's
	// mode clause.
	purposeOnMode
	// purposeOnSpellMode is a bullet of a modal spell (Spec.Modes). It
	// may declare Answers, which a cast choosing it reads (#2877); an
	// ability's bullet may not, since its row declares for it.
	purposeOnSpellMode
	purposeOnAltCost
	purposeOnActivated
	purposeOnAnyPlayerActivated
	purposeOnTriggered
	// purposeOnFaceUp is the inner purpose of Purpose.FaceUp (#3776): the
	// card's own slot, plus the one-turn pump a turn-face-up trigger gives
	// its source ("it gets +6/+6 and gains trample until end of turn").
	purposeOnFaceUp
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
	if p.LandEntry != "" {
		fail("sets LandEntry by hand; the catalog replacement projection owns it (#4317)")
	}
	if p.ReturnsLandOnEntry && slot != purposeOnCard {
		fail("sets ReturnsLandOnEntry off the card (#4317)")
	}
	if p.Draws < 0 || p.ControllerLosesLife < 0 || p.Discards < 0 || p.Lands < 0 ||
		p.Tutors < 0 || p.SelfMillTutor < 0 || p.Tokens < 0 || p.Energy < 0 || p.Sweep.Amount < 0 ||
		p.ExtraCombat < 0 || p.ExtraTurns < 0 || p.SkipsTurns < 0 || p.DamageToCreature < 0 || p.DamageEachOpponent < 0 || p.LifeGain < 0 ||
		p.AwakenLand < 0 || p.ExtraLandDrops < 0 || p.LandfallExtraTriggers < 0 || p.ExtraLoyaltyActivations < 0 || p.LandsUntapped < 0 || p.Impulse < 0 || p.CreatureTokens < 0 || p.SourceCounters < 0 {
		fail("has a negative amount")
	}
	if p.LandsUntapped > p.Lands {
		fail("says more lands enter untapped (LandsUntapped) than it puts onto the battlefield (Lands) (ADR 0136 §2)")
	}
	if p.Aura != "" && slot != purposeOnCard {
		fail("sets Aura off the card — whether an Aura helps or hurts its host is the card's own (#3309)")
	}
	if p.Aura != "" && p.Aura != game.AuraBenefit && p.Aura != game.AuraHarm {
		fail("sets an Aura other than benefit or harm (#3309)")
	}
	if p.LandfallExtraTriggers != 0 && slot != purposeOnCard {
		fail("sets LandfallExtraTriggers off the card — a landfall doubler is the card's own static (#3146)")
	}
	if p.ExtraLoyaltyActivations != 0 && slot != purposeOnCard && slot != purposeOnActivated {
		fail("sets ExtraLoyaltyActivations off the card or an activated row of its controller's own — it grants the planeswalkers of whoever resolves it (#4191)")
	}
	if p.SourceCounters != 0 && slot != purposeOnActivated {
		fail("sets SourceCounters off an activated row of its controller's own — it is the +1/+1 counters a row puts on its own source, and an any-player row grows another player's permanent (#5083)")
	}
	if p.ExtraLandDrops != 0 && slot != purposeOnCard {
		fail("sets ExtraLandDrops off the card — an additional land drop is the card's own static or its spell's effect (#2678)")
	}
	if p.ControllerLosesLife != 0 && slot != purposeOnAnyPlayerActivated {
		fail("names ControllerLosesLife off an any-player activated row — the activator and the controller are one player there")
	}
	if (p.ExtraTurns != 0 || p.SkipsTurns != 0) && slot == purposeOnAnyPlayerActivated {
		fail("sets ExtraTurns or SkipsTurns on an any-player activated row — the activator is not the controller, so the caster's own turns would be the wrong player's, and a bot must never activate another player's copy (ADR 0106 decision 2, ADR 0188 Decision 10)")
	}
	if p.DeathPayoff && slot != purposeOnTriggered {
		fail("sets DeathPayoff off a triggered row")
	}
	if p.DiscardPayoff != nil && slot != purposeOnTriggered {
		fail("sets DiscardPayoff off a triggered row")
	}
	if p.OnOpponentDiscard != nil && slot != purposeOnTriggered {
		fail("sets OnOpponentDiscard off a triggered row")
	}
	if p.LandfallPayoff != nil && slot != purposeOnTriggered {
		fail("sets LandfallPayoff off a triggered row — a landfall payoff is a \"whenever a land you control enters\" trigger (#3072)")
	}
	if p.CreatureEnterPayoff != nil && slot != purposeOnTriggered {
		fail("sets CreatureEnterPayoff off a triggered row — it is a \"whenever a creature you control enters\" trigger (#3165)")
	}
	if p.LifeGainPayoff != nil && slot != purposeOnTriggered {
		fail("sets LifeGainPayoff off a triggered row — it is a \"whenever you gain life\" trigger (#4979)")
	}
	if p.CreatureTokens > p.Tokens {
		fail("declares more CreatureTokens than Tokens (#3165)")
	}
	if p.TokensPerControlled != "" && !onActivated(slot) && slot != purposeOnTriggered {
		fail("sets TokensPerControlled off a triggered or activated row — it counts at resolution (#3165)")
	}
	if t := p.TokensPerControlled; t != strings.ToLower(t) || strings.ContainsAny(t, " \t—") {
		fail("names a TokensPerControlled subtype that is not one lowercase word (#3165)")
	}
	onSource := slot == purposeOnTriggered || slot == purposeOnFaceUp || slot == purposeOnActivated || slot == purposeOnAnyPlayerActivated
	if p.Pump != nil && !onSource {
		fail("sets Pump off a triggered or activated row — a pump is what an ability gives its own source")
	}
	if p.Pump != nil && p.Pump.CantBlock {
		fail("sets CantBlock on a row's pump of its own source — \"can't block\" is a harm, declared on a target's entry (#4977)")
	}
	if p.PreventCombatDamageToSelf && !onSource {
		fail("sets PreventCombatDamageToSelf off a triggered or activated row")
	}
	if p.Answers != 0 && !answersSlot(slot) && !(slot == purposeOnTriggered && p.Answers == game.AnswerRemove) {
		fail("declares Answers off an activated row, the card, a spell's mode or an alternative cost, other than AnswerRemove (an edict) on a triggered one — nothing reads it there (ADR 0142 §2 and its amendment of 2026-10-30, ADR 0009 #3882)")
	}
	checkAnswerBits(p.Answers, fail)
	checkSaves(p, slot, fail)
	if p.AwakenLand != 0 && slot != purposeOnAltCost {
		fail("sets AwakenLand off an alternative cost — it is what an awaken offer adds (ADR 0135 §3)")
	}
	checkPump(p.Pump, fail)
	checkDiscardPayoff(p.DiscardPayoff, fail)
	checkDiscardPayoff(p.OnOpponentDiscard, fail)
	checkLandfallPayoff(p.LandfallPayoff, fail)
	if c := p.CreatureEnterPayoff; c != nil {
		l := game.LandfallPayoff(*c)
		checkLandfallPayoff(&l, fail)
	}
	checkLifeGainPayoff(p.LifeGainPayoff, fail)
	checkSweep(p.Sweep, fail)
	checkRiders(p, slot, fail)
	checkResolutionMana(p.ResolutionMana, slot, fail)
	if (p.Kicked != nil || p.FaceUp != nil) && slot != purposeOnCard {
		fail("sets Kicked or FaceUp off the card — what a kicker or a turn face up adds is the card's own (#3762)")
	}
}

// checkResolutionMana is the guard for Purpose.ResolutionMana (#4980):
// the mana an activated row of its controller's own adds on one
// numbered resolution of it in a turn (Soulbright Seeker's third). It is
// refused off such a row, on a resolution before the first, and on mana
// that is not one or more fixed symbols ({W}{U}{B}{R}{G}{C}): a generic
// number, an X, a hybrid or a Phyrexian symbol would let the reader
// count mana the card does not add.
func checkResolutionMana(rm *game.ResolutionMana, slot purposeSlot, fail func(string)) {
	if rm == nil {
		return
	}
	if slot != purposeOnActivated {
		fail("sets ResolutionMana off an activated row of its controller's own — it is the mana a row adds on its nth resolution in a turn (#4980)")
	}
	if rm.Resolution < 1 {
		fail("sets ResolutionMana with a resolution before the first (#4980)")
	}
	notFixed := fmt.Sprintf("sets ResolutionMana's mana to %q, which is not one or more fixed mana symbols (#4980)", rm.Mana)
	c, err := game.ParseCost(rm.Mana)
	if err != nil || len(c.Required) == 0 || c.Generic != 0 || c.XSlots != 0 {
		fail(notFixed)
	}
	for _, r := range c.Required {
		if len(r.Options) != 1 || r.Snow || r.Phyrexian || r.HasNumericAlt {
			fail(notFixed)
		}
	}
}

// checkRiders is the guard for the riders (#4978, ADR 0126's amendment
// of 2026-11-05): scry, learn, discover, amass and the Ring tempting
// its controller. None may be negative, and none is declared on a
// triggered row, which nothing prices by them: a permanent's enters
// trigger declares its rider on the card, as Mulldrifter's draws are,
// and a modal trigger's bullet on that bullet.
func checkRiders(p game.Purpose, slot purposeSlot, fail func(string)) {
	if p.Scry < 0 || p.Learn < 0 || p.Discover < 0 || p.Amass < 0 || p.RingTempts < 0 {
		fail("has a negative rider (#4978)")
	}
	if slot == purposeOnTriggered && (p.Scry != 0 || p.Learn != 0 || p.Discover != 0 || p.Amass != 0 || p.RingTempts != 0) {
		fail("declares a rider (scry, learn, discover, amass, the Ring tempting) on a triggered row, which nothing reads — an enters trigger's rider is the card's own purpose (#4978)")
	}
}

// onActivated reports whether a slot is an activated row.
func onActivated(slot purposeSlot) bool {
	return slot == purposeOnActivated || slot == purposeOnAnyPlayerActivated
}

// answersSlot reports whether Answers may be declared in a slot: an
// activated row (ADR 0142 §2), and since #2877 the card, a spell's mode
// and an alternative cost, which a cast reads (ADR 0142's amendment of
// 2026-10-30; internal/legal's castAnswersOf).
func answersSlot(slot purposeSlot) bool {
	switch slot {
	case purposeOnCard, purposeOnSpellMode, purposeOnAltCost:
		return true
	}
	return onActivated(slot)
}

// checkSaves is checkPurpose's half for a declared Saves (#2939 gap 3):
// on an activated row that declares survive or protect, naming one of
// the sweep classes.
func checkSaves(p game.Purpose, slot purposeSlot, fail func(string)) {
	s := p.Saves
	if s == nil {
		return
	}
	if !onActivated(slot) {
		fail("sets Saves off an activated row — it is what a survive or protect row covers")
	}
	if !p.Answers.HasAny(game.AnswerSurvive | game.AnswerProtect) {
		fail("sets Saves on a row that declares neither survive nor protect")
	}
	if !slices.Contains(game.SweepMatches, s.Matches) {
		fail(fmt.Sprintf("sets Saves to the unknown class %q", s.Matches))
	}
}

// checkPump is checkPurpose's half for a declared Pump (ADR 0130's
// amendment of 2026-10-07): it gives something, and each keyword is one
// lowercase word, as the wire's `abilities` spells it.
func checkPump(pm *game.Pump, fail func(string)) {
	if pm == nil {
		return
	}
	if pm.Power == 0 && pm.Toughness == 0 && len(pm.Keywords) == 0 && !pm.CantBeBlocked && !pm.CantBlock {
		fail("is a pump that gives nothing — set Power, Toughness, Keywords, CantBeBlocked or CantBlock")
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

// checkLandfallPayoff is checkPurpose's half for a declared
// LandfallPayoff (#3072): no negative amount, and it pays something.
func checkLandfallPayoff(l *game.LandfallPayoff, fail func(string)) {
	if l == nil {
		return
	}
	if l.Draws < 0 || l.LifeGain < 0 || l.Tokens < 0 || l.Mana < 0 || l.Counters < 0 || l.Energy < 0 || l.DamageEachOpponent < 0 {
		fail("has a negative landfall amount")
	}
	if *l == (game.LandfallPayoff{}) {
		fail("is a landfall payoff that pays nothing — leave it nil, or set what each land pays")
	}
}

// checkLifeGainPayoff is checkPurpose's half for a declared
// LifeGainPayoff (#4979): checkLandfallPayoff's rules, and no LifeGain,
// which would be the row paying for its own payout.
func checkLifeGainPayoff(g *game.LifeGainPayoff, fail func(string)) {
	if g == nil {
		return
	}
	l := game.LandfallPayoff(*g)
	checkLandfallPayoff(&l, fail)
	if g.LifeGain != 0 {
		fail("is a lifegain payoff that gains life — it would trigger itself; leave LifeGain out (#4979)")
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
// a mode slot whatever owns it: a spell (purposeOnSpellMode), an
// activated or a triggered ability (purposeOnMode).
func checkModePurposes(name, where string, slot purposeSlot, m *game.ModeSpec) {
	if m == nil {
		return
	}
	for i, o := range m.Options {
		mw := fmt.Sprintf("%s mode %d", where, i)
		checkPurpose(name, mw, slot, o.Purpose)
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
	if spec.Purpose.LandfallExtraTriggers != 0 && len(spec.TriggerDoublers) == 0 {
		panic(fmt.Sprintf("effects.Register: %q declares LandfallExtraTriggers but has no TriggerDoubler: the purpose says what the static does (#3146)", name))
	}
	checkTargetPurposes(name, "card", spec.Purpose, spec.Targets, true)
	checkConditionalPurposes(spec)
	checkModePurposes(name, "spell", purposeOnSpellMode, spec.Modes)
	checkSpellAnswers(spec)
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
	instantLoyalty := ownLoyaltyAtInstant(spec.ActivationTimings)
	for i, a := range spec.Activated {
		checkActivatedPurpose(name, fmt.Sprintf("ability %d", i), a, instantLoyalty)
	}
	for i, t := range spec.Triggered {
		checkTriggeredPurpose(name, fmt.Sprintf("triggered ability %d", i), t)
	}
	checkManaAbilitiesAnswers(name, "card", spec.ManaAbilities)
}

// checkConditionalPurposes is the guard for Purpose.Kicked and
// Purpose.FaceUp (ADR 0126's amendment of 2026-10-15, #3762). Each is
// read only when its condition holds, so it must be one a reader can
// recognise:
//
//   - Kicked on a card with exactly one optional cost, which is a
//     (single) kicker. With two kickers ("Kicker {R} and/or {W}") the
//     view cannot say which was paid, and a multikicker's amounts are
//     counted at resolution; both keep their old, unconditional silence.
//   - FaceUp on a card with a face-down cast offer (morph, megamorph,
//     disguise), the only way the special action exists to price.
//
// The inner purpose says something, and is a plain one: no further
// Kicked or FaceUp, no Targets (the clauses it would index belong to
// the statement the cast announces), and nothing that is the card
// slot's own static (Aura, ExtraLandDrops, LandfallExtraTriggers) or
// the view's own stamp (FlashCleanupSacrifice).
func checkConditionalPurposes(spec Spec) {
	name := spec.Name
	inner := func(which string, p *game.Purpose) {
		if p == nil {
			return
		}
		where := "card " + which
		fail := func(why string) {
			panic(fmt.Sprintf("effects.Register: %q %s declares a Purpose that %s (#3762)", name, where, why))
		}
		switch {
		case p.IsZero():
			fail("says nothing — leave it nil")
		case p.Kicked != nil || p.FaceUp != nil:
			fail("nests Kicked or FaceUp")
		case p.Targets != nil:
			fail("declares Targets — a conditional purpose has no clause list of its own")
		case p.Aura != "" || p.ExtraLandDrops != 0 || p.LandfallExtraTriggers != 0 || p.FlashCleanupSacrifice:
			fail("declares Aura, ExtraLandDrops, LandfallExtraTriggers or FlashCleanupSacrifice, which are the card's own")
		case p.Answers != 0:
			// #2877: what a spell answers is declared on the card, as the
			// union of everything it can do (ADR 0142 sweep ruling 4), so
			// a kicked cast is read like any other. A turn face up is a
			// special action, which nothing asks.
			fail("declares Answers — declare what the card answers on the card, the union of its kicked and unkicked effect (ADR 0142's amendment of 2026-10-30)")
		}
		slot := purposeOnCard
		if which == "face up" {
			slot = purposeOnFaceUp
		}
		checkPurpose(name, where, slot, *p)
	}
	if spec.Purpose.Kicked != nil {
		if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != game.KickerKey {
			panic(fmt.Sprintf("effects.Register: %q declares Purpose.Kicked but not exactly one kicker optional cost — "+
				"with two kickers the view cannot say which was paid (#3762)", spec.Name))
		}
		inner("kicked", spec.Purpose.Kicked)
	}
	if spec.Purpose.FaceUp != nil {
		offered := false
		for _, a := range spec.AlternativeCosts {
			offered = offered || a.FaceDown != nil
		}
		if !offered {
			panic(fmt.Sprintf("effects.Register: %q declares Purpose.FaceUp but has no face-down cast offer (Morph, Megamorph, Disguise) (#3762)", spec.Name))
		}
		inner("face up", spec.Purpose.FaceUp)
	}
}

// checkActivatedPurpose checks one activated row's purpose and its
// modes'. Shared with the ability-grant guard, because a granted row
// is declared exactly as a card's own is. instantLoyalty is
// checkActivatedAnswers'.
func checkActivatedPurpose(name, where string, a ActivatedAbility, instantLoyalty bool) {
	slot := purposeOnActivated
	if a.AnyPlayer {
		slot = purposeOnAnyPlayerActivated
	}
	checkPurpose(name, where, slot, a.Purpose)
	checkActivatedAnswers(name, where, activatedShapeOf(a), instantLoyalty)
	checkTargetPurposes(name, where, a.Purpose, a.Targets, true)
	checkModePurposes(name, where, purposeOnMode, a.Modes)
}

// checkTriggeredPurpose checks one triggered row's purpose and its
// modes'. Shared with the ability-grant guard. A row whose clause list
// is built at trigger time (TargetsFrom) has no statement to check a
// target entry against, so it may not declare one.
func checkTriggeredPurpose(name, where string, t game.TriggeredAbility) {
	checkPurpose(name, where, purposeOnTriggered, t.Purpose)
	checkTargetPurposes(name, where, t.Purpose, t.Targets, t.Targets != nil || t.TargetsFrom == nil)
	checkModePurposes(name, where, purposeOnMode, t.Modes)
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
//     is the rest of the review;
//   - a change until end of turn (#4302) that is neither benefit nor
//     harm, on a clause that cannot target a permanent, or whose Pump
//     points the other way (checkUntilEndOfTurn); a pump of the
//     announced X (#4976) on a harm, or beside a printed amount on the
//     same side (checkPumpX);
//   - an exile from a graveyard (#4559) on a clause that cannot target
//     a graveyard card, beside a reanimation, or an ExileGrows that is
//     not "permanent" or "creature" or has no Exiles beside it.
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
		if e.Draws < 0 || e.Discards < 0 || e.ExtraTurns < 0 || e.Tokens < 0 || e.LifeGain < 0 || e.LifeLoss < 0 || e.Damage < 0 || e.DamagePerTapped < 0 {
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
		if e.DamagePerTapped != 0 && (e.Damage != 0 || e.DamageIsX) {
			fail(fmt.Sprintf("sets DamagePerTapped beside Damage or DamageIsX on slot %d", e.Slot))
		}
		if (e.Damage != 0 || e.DamageIsX || e.DamagePerTapped != 0) && !clause.Players && !slices.Contains(clause.Zones, game.ZoneBattlefield) {
			fail(fmt.Sprintf("deals damage to slot %d, whose clause %q can target neither a player nor a permanent", e.Slot, clause.Label))
		}
		if e.Reanimates && !slices.Contains(clause.Zones, game.ZoneGraveyard) {
			fail(fmt.Sprintf("reanimates slot %d, whose clause %q cannot target a card in a graveyard", e.Slot, clause.Label))
		}
		if e.Exiles && !slices.Contains(clause.Zones, game.ZoneGraveyard) {
			fail(fmt.Sprintf("exiles slot %d from a graveyard, whose clause %q cannot target a card in a graveyard (#4559)", e.Slot, clause.Label))
		}
		if e.Exiles && e.Reanimates {
			fail(fmt.Sprintf("both exiles and reanimates slot %d (#4559)", e.Slot))
		}
		switch e.ExileGrows {
		case "":
		case game.ExileGrowsPermanent, game.ExileGrowsCreature:
			if !e.Exiles {
				fail(fmt.Sprintf("sets ExileGrows on slot %d without Exiles (#4559)", e.Slot))
			}
		default:
			fail(fmt.Sprintf("sets ExileGrows %q on slot %d, which is neither %q nor %q (#4559)", e.ExileGrows, e.Slot, game.ExileGrowsPermanent, game.ExileGrowsCreature))
		}
		checkUntilEndOfTurn(e, clause, fail)
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

// checkUntilEndOfTurn is checkTargetPurposes' half for a change that
// ends with the turn (#4302): a direction of benefit or harm, on a
// clause that can pick a permanent; a Pump only beside one, that says
// something, adds nothing negative to a benefit and nothing positive or
// no keyword to a harm (a harm's keywords are lost, which a Pump cannot
// say).
func checkUntilEndOfTurn(e game.TargetPurpose, clause *game.TargetClause, fail func(string)) {
	checkPumpX(e, fail)
	switch e.UntilEndOfTurn {
	case "":
		if e.Pump != nil {
			fail(fmt.Sprintf("sets a Pump on slot %d without UntilEndOfTurn — say whether it helps or hurts (#4302)", e.Slot))
		}
		return
	case game.TargetBenefit, game.TargetHarm:
	default:
		fail(fmt.Sprintf("sets UntilEndOfTurn %q on slot %d, which is neither benefit nor harm (#4302)", e.UntilEndOfTurn, e.Slot))
	}
	if !slices.Contains(clause.Zones, game.ZoneBattlefield) {
		fail(fmt.Sprintf("changes slot %d until end of turn, whose clause %q cannot target a permanent (#4302)", e.Slot, clause.Label))
	}
	pm := e.Pump
	if pm == nil {
		return
	}
	checkPump(pm, fail)
	if e.UntilEndOfTurn == game.TargetBenefit && (pm.Power < 0 || pm.Toughness < 0) {
		fail(fmt.Sprintf("declares a benefit on slot %d that takes power or toughness — that is a harm (#4302)", e.Slot))
	}
	if e.UntilEndOfTurn == game.TargetHarm && (pm.Power > 0 || pm.Toughness > 0 || len(pm.Keywords) > 0 || pm.CantBeBlocked) {
		fail(fmt.Sprintf("declares a harm on slot %d that adds power, toughness, a keyword or unblockability (#4302)", e.Slot))
	}
	if e.UntilEndOfTurn == game.TargetBenefit && pm.CantBlock {
		fail(fmt.Sprintf("declares a benefit on slot %d that says it can't block — that is a harm (#4977)", e.Slot))
	}
}

// checkPumpX is checkUntilEndOfTurn's half for a pump of the announced X
// (#4976): only on a benefit, and with no printed amount on the side the
// X fills, as DamageIsX is refused beside Damage.
func checkPumpX(e game.TargetPurpose, fail func(string)) {
	if !e.PumpPowerIsX && !e.PumpToughnessIsX {
		return
	}
	if e.UntilEndOfTurn != game.TargetBenefit {
		fail(fmt.Sprintf("declares a pump of X on slot %d that is not a benefit until end of turn (#4976)", e.Slot))
	}
	if pm := e.Pump; pm != nil && ((e.PumpPowerIsX && pm.Power != 0) || (e.PumpToughnessIsX && pm.Toughness != 0)) {
		fail(fmt.Sprintf("sets both a printed amount and X on the same side of slot %d's pump (#4976)", e.Slot))
	}
}

// PumpXUntilEOT is BenefitUntilEOT for "+X/+X until end of turn", where
// X is the spell's or ability's announced X (#4976): Tyvar's Stand's
// "target creature you control gets +X/+X and gains hexproof and
// indestructible until end of turn" is PumpXUntilEOT(0, "hexproof",
// "indestructible"). The bot reads the X the move names.
func PumpXUntilEOT(slot int, keywords ...string) game.TargetPurpose {
	e := BenefitUntilEOT(slot, 0, 0, keywords...)
	e.PumpPowerIsX, e.PumpToughnessIsX = true, true
	return e
}

// PowerXUntilEOT is PumpXUntilEOT for "+X/+0" (#4976): Hatred's "target
// creature gets +X/+0 until end of turn" is PowerXUntilEOT(0), and Kessig
// Wolf Run's "+X/+0 and gains trample" PowerXUntilEOT(0, "trample").
func PowerXUntilEOT(slot int, keywords ...string) game.TargetPurpose {
	e := BenefitUntilEOT(slot, 0, 0, keywords...)
	e.PumpPowerIsX = true
	return e
}

// BenefitUntilEOT is the target entry for a clause that helps its pick
// until end of turn (#4302): Heliod, Sun-Crowned's "another target
// creature gains lifelink until end of turn" is
// ForTargets(BenefitUntilEOT(0, 0, 0, "lifelink")), and "target creature
// gets +2/+2 until end of turn" BenefitUntilEOT(0, 2, 2). With no amount
// and no keyword (+X/+X where X is counted at resolution, "can't be
// blocked this turn") it declares only the direction; an announced X is
// PumpXUntilEOT or PowerXUntilEOT (#4976).
func BenefitUntilEOT(slot, power, toughness int, keywords ...string) game.TargetPurpose {
	e := game.TargetPurpose{Slot: slot, UntilEndOfTurn: game.TargetBenefit}
	if power != 0 || toughness != 0 || len(keywords) > 0 {
		e.Pump = &game.Pump{Power: power, Toughness: toughness, Keywords: keywords}
	}
	return e
}

// CantBeBlockedUntilEOT is BenefitUntilEOT for a clause that also makes
// its pick unblockable this turn (#4312): Rogue's Passage's "target
// creature can't be blocked this turn" is CantBeBlockedUntilEOT(0, 0, 0),
// and Taigam's Strike's "+2/+0 until end of turn and can't be blocked this
// turn" CantBeBlockedUntilEOT(0, 2, 0).
func CantBeBlockedUntilEOT(slot, power, toughness int, keywords ...string) game.TargetPurpose {
	e := BenefitUntilEOT(slot, power, toughness, keywords...)
	if e.Pump == nil {
		e.Pump = &game.Pump{}
	}
	e.Pump.CantBeBlocked = true
	return e
}

// HarmUntilEOT is the target entry for a clause that hurts its pick until
// end of turn (#4302), with the printed change as signed amounts: "target
// creature gets -2/-1 until end of turn" is HarmUntilEOT(0, -2, -1). With
// no amount (-X/-X) it declares only the direction. "Can't block this
// turn" is CantBlockUntilEOT (#4977).
func HarmUntilEOT(slot, power, toughness int) game.TargetPurpose {
	e := game.TargetPurpose{Slot: slot, UntilEndOfTurn: game.TargetHarm}
	if power != 0 || toughness != 0 {
		e.Pump = &game.Pump{Power: power, Toughness: toughness}
	}
	return e
}

// CantBlockUntilEOT is HarmUntilEOT for a clause that also says its pick
// can't block this turn (#4977): Daring Discovery's "up to three target
// creatures can't block this turn" is CantBlockUntilEOT(0, 0, 0), and
// "target creature gets -2/-0 and can't block this turn"
// CantBlockUntilEOT(0, -2, 0). With Config.PriceCantBlock on the bot
// prices it by the attack it opens before the defender's blocks, not as
// removal.
func CantBlockUntilEOT(slot, power, toughness int) game.TargetPurpose {
	e := HarmUntilEOT(slot, power, toughness)
	if e.Pump == nil {
		e.Pump = &game.Pump{}
	}
	e.Pump.CantBlock = true
	return e
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

// DamagePerTappedToTarget is the target entry for "deals damage to" the
// pick of clause `slot` equal to `per` times the number of creatures the
// spell's "tap any number" cost tapped (#3961): Burn at the Stake is
// ForTargets(DamagePerTappedToTarget(0, 3)). The bot reads the count the
// move names, so tapping more is worth what the damage buys.
func DamagePerTappedToTarget(slot, per int) game.TargetPurpose {
	return game.TargetPurpose{Slot: slot, DamagePerTapped: per}
}

// Reanimates is the target entry for a spell that puts the graveyard
// card picked for clause `slot` onto the battlefield under its caster's
// control (#3161): Reanimate is ForTargets(Reanimates(0)). The bot prices
// the pick by the permanent it becomes, whoever owns the graveyard.
func Reanimates(slot int) game.TargetPurpose {
	return game.TargetPurpose{Slot: slot, Reanimates: true}
}

// ExilesFromGraveyard is the target entry for an effect that exiles the
// graveyard card picked for clause `slot` (#4559). `grows` is the card
// type whose exile puts a +1/+1 counter on the source
// (game.ExileGrowsPermanent or game.ExileGrowsCreature), or "" when it
// puts none: Lion Sash is
// ForTargets(ExilesFromGraveyard(0, game.ExileGrowsPermanent)). The bot
// prices the pick by what the exile denies an opponent or costs the bot,
// plus the counter, rather than as a card it could use.
func ExilesFromGraveyard(slot int, grows string) game.TargetPurpose {
	return game.TargetPurpose{Slot: slot, Exiles: true, ExileGrows: grows}
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

// WithRiders is p with the riders r declares added (#4978): scry,
// learn, discover, amass and the Ring tempting its controller, for a
// purpose a constructor builds. Only r's riders are read.
//
//	Purpose: WithRiders(ForTargets(DamageToTarget(0, 4)), game.Purpose{Scry: 1}),
func WithRiders(p, r game.Purpose) game.Purpose {
	p.Scry += r.Scry
	p.Learn += r.Learn
	p.Amass += r.Amass
	p.RingTempts += r.RingTempts
	if r.Discover != 0 {
		p.Discover = r.Discover
	}
	return p
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

// TriggerEdict declares a triggered row an untargeted edict ("each
// opponent sacrifices a creature"): it sets AnswerRemove on the row's
// Purpose, keeping what it already declares. Smart autopass reads it,
// with a declared Sweep, when an opponent's trigger is on the stack
// (ADR 0009's #3882 amendment). A sweep needs no wrapper: its
// Purpose.Sweep is read as it stands.
//
//	TriggerEdict(WhenThisEnters("Fleshbag Marauder — each player sacrifices a creature", ...))
func TriggerEdict(t game.TriggeredAbility) game.TriggeredAbility {
	t.Purpose.Answers |= game.AnswerRemove
	return t
}

// LandfallPaying is Landfall with what the row pays for each land
// declared (#3072): the purpose the heuristic reads when it prices a
// land the bot puts onto the battlefield. Tatyova is
//
//	LandfallPaying("Tatyova — gain 1 life and draw a card (landfall)",
//		game.LandfallPayoff{Draws: 1, LifeGain: 1}, effect)
//
// A row whose payoff is counted at resolution or depends on the board
// (Avenger of Zendikar's counters, Field of the Dead's Zombie) is a
// plain Landfall: no printed amount would be true.
func LandfallPaying(label string, pays game.LandfallPayoff, effect Effect) game.TriggeredAbility {
	return TriggerWithPurpose(Landfall(label, effect), game.Purpose{LandfallPayoff: &pays})
}

// CreatureEnterPaying is a triggered row with what it pays for each
// creature that enters under its controller's control declared (#3165):
// the purpose the heuristic reads when it prices a token maker.
// Purphoros is
//
//	CreatureEnterPaying(WheneverAnotherCreatureEntersUnderYourControl(...),
//		game.CreatureEnterPayoff{DamageEachOpponent: 2})
func CreatureEnterPaying(t game.TriggeredAbility, pays game.CreatureEnterPayoff) game.TriggeredAbility {
	return TriggerWithPurpose(t, game.Purpose{CreatureEnterPayoff: &pays})
}

// LifeGainPaying is a "whenever you gain life" row with what it pays for
// each life-gain event declared (#4979): the purpose the heuristic reads
// when it prices a lifelink attack. Ajani's Pridemate is
//
//	LifeGainPaying(WheneverYouGainLife(...), game.LifeGainPayoff{Counters: 1})
func LifeGainPaying(t game.TriggeredAbility, pays game.LifeGainPayoff) game.TriggeredAbility {
	return TriggerWithPurpose(t, game.Purpose{LifeGainPayoff: &pays})
}
