package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// answers.go — ADR 0142: the registration guard for a declared
// Purpose.Answers on an activated row, ManaAbility.Answers on a mana
// ability, and (#2877, ADR 0142's amendment of 2026-10-30) Answers on a
// spell, a spell's mode or an alternative cost. The vocabulary and its
// tiers are game/answers.go; the reader is internal/legal's
// untargetedFlags (manaMoveInteracts for a mana ability, castAnswersOf
// for a cast).
//
// A declaration is the only thing the reader reads, so the guard keeps
// it honest where it can without reading text. It refuses:
//
//   - a bit that names no answer;
//   - AnswerValue beside any other answer;
//   - Answers on an ability's mode, inside Kicked or FaceUp, or anything
//     but AnswerRemove (an edict) on a triggered row, which nothing
//     reads it on (checkPurpose, ADR 0142 §2);
//   - on a spell, `value` in any slot when its additional cost
//     sacrifices or exiles a creature, or on an alternative cost that
//     sacrifices or returns one (checkSpellAnswers);
//   - Answers on a row whose every announcement must name a target (a
//     clause with Min 0, "up to one", still has the empty announcement):
//     its move carries has_targets, which already stops smart autopass;
//   - Answers on a sorcery-speed row, which is never offered while an
//     item waits on the stack. A loyalty row is one, unless its card
//     opens its own loyalty abilities at instant speed (Teferi, Master
//     of Time; game.ActivationTiming.OwnLoyalty);
//   - a crew row (AbilityCost.Crew) that does not declare AnswerAnimate;
//   - a row whose cost sacrifices, exiles or returns a creature
//     (legal.CostAnswers says sac_outlet) that declares neither
//     AnswerSacOutlet, AnswerProtect nor AnswerSurvive, so the #2853
//     cost rule stays true for every declared row (ADR 0142 §3);
//   - on a mana ability, anything but AnswerSacOutlet or AnswerValue, or
//     a creature-sacrifice cost declared without AnswerSacOutlet.

// checkAnswerBits is the part of the guard every slot shares: known
// bits, and value alone.
func checkAnswerBits(a game.Answers, fail func(string)) {
	if a == 0 {
		return
	}
	if u := a.Unknown(); u != 0 {
		fail(fmt.Sprintf("declares Answers bits %#x that name no answer", uint16(u)))
	}
	if a.HasAny(game.AnswerValue) && a != game.AnswerValue {
		fail(fmt.Sprintf("declares AnswerValue beside other answers (%s) — value stands alone", a))
	}
}

// checkActivatedAnswers is the guard for one activated row's declared
// answers. Shared by a card's own rows, a granted bundle's and a token
// template's. instantLoyalty is true for a card's own rows when the card
// opens its own loyalty abilities at instant speed (ownLoyaltyAtInstant).
func checkActivatedAnswers(name, where string, ab game.ActivatedAbilityShape, instantLoyalty bool) {
	a := ab.Purpose.Answers
	if a == 0 {
		return
	}
	fail := func(why string) {
		panic(fmt.Sprintf("effects.Register: %q %s declares Answers that %s (ADR 0142; docs/adding-cards.md, \"Declaring what an ability answers\")", name, where, why))
	}
	checkAnswerBits(a, fail)
	if sorcerySpeedRow(ab, instantLoyalty) {
		fail("are on a sorcery-speed or loyalty row, which is never offered while an item waits on the stack")
	}
	if !hasUntargetedAnnouncement(ab) {
		fail("are on a row whose every announcement has a target — its move carries has_targets, which already stops smart autopass")
	}
	if ab.Cost.Crew > 0 && !a.Has(game.AnswerAnimate) {
		fail("leave out AnswerAnimate on a crew row — crewing makes the Vehicle a creature")
	}
	if legal.CostAnswers(ab.Cost).Has(game.AnswerSacOutlet) && !a.HasAny(game.AnswerSacOutlet|game.AnswerProtect|game.AnswerSurvive) {
		fail(fmt.Sprintf("(%s) leave out AnswerSacOutlet, AnswerProtect and AnswerSurvive on a row whose cost sacrifices, exiles or returns a creature", a))
	}
}

// sorcerySpeedRow reports whether the row is only ever activated as a
// sorcery: "activate only as a sorcery" (CR 602.5d), or a loyalty
// ability (CR 606.3) on a card that does not open its own loyalty
// abilities at instant speed (instantLoyalty).
func sorcerySpeedRow(ab game.ActivatedAbilityShape, instantLoyalty bool) bool {
	return ab.SorcerySpeed || (ab.Cost.Loyalty != nil && !instantLoyalty)
}

// ownLoyaltyAtInstant reports whether a card's activation-timing
// statements open its own loyalty abilities at instant speed, so its
// loyalty rows can be offered while an opponent's item waits on the
// stack and smart autopass reads what they answer. A statement that
// opens every planeswalker's (Teferi, Temporal Archmage's emblem) is
// not one: it is about other cards' rows.
func ownLoyaltyAtInstant(timings []game.ActivationTiming) bool {
	for _, t := range timings {
		if t.OwnLoyalty && t.Grants() {
			return true
		}
	}
	return false
}

// answersInScope reports whether smart autopass reads the row's answers:
// it may be activated at instant speed with no target (ADR 0142 §4).
// TestEveryAnswersRowDeclares fails on such a catalog row that declares
// nothing.
func answersInScope(ab game.ActivatedAbilityShape, instantLoyalty bool) bool {
	return !sorcerySpeedRow(ab, instantLoyalty) && hasUntargetedAnnouncement(ab)
}

// hasUntargetedAnnouncement reports whether the row can be announced
// with no target: it has no target clause (or one that allows zero
// targets) and, if it is modal, at least one mode with none.
func hasUntargetedAnnouncement(ab game.ActivatedAbilityShape) bool {
	if targetsRequired(ab.Targets) {
		return false
	}
	if ab.Modes == nil {
		return true
	}
	for _, o := range ab.Modes.Options {
		if !targetsRequired(o.Targets) {
			return true
		}
	}
	return false
}

// targetsRequired reports whether a target clause forces an announcement
// to name a target. A nil clause does not, and neither does one whose
// Min is 0 ("up to one target", "any number of"): the enumerator offers
// the empty announcement beside the targeted ones (ADR 0142, amendment
// of 2026-10-10).
func targetsRequired(t *game.TargetSpec) bool {
	return t != nil && (t.Min > 0 || t.CountFromX)
}

// checkManaAnswers is the guard for a mana ability's declared answers.
func checkManaAnswers(name, where string, m game.ManaAbilityShape) {
	a := m.Answers
	if a == 0 {
		return
	}
	fail := func(why string) {
		panic(fmt.Sprintf("effects.Register: %q %s declares Answers that %s (ADR 0142 §2)", name, where, why))
	}
	checkAnswerBits(a, fail)
	if a != game.AnswerValue && a != game.AnswerSacOutlet {
		fail(fmt.Sprintf("(%s) are neither AnswerSacOutlet nor AnswerValue — a mana ability answers only as a sacrifice outlet", a))
	}
	if legal.ManaCostAnswers(m.SacrificeOther).Has(game.AnswerSacOutlet) && a != game.AnswerSacOutlet {
		fail("leave out AnswerSacOutlet on a mana ability whose cost sacrifices a creature")
	}
}

// activatedShapeOf is the part of a declared ActivatedAbility the
// answers guard reads, as the engine's shape.
func activatedShapeOf(a ActivatedAbility) game.ActivatedAbilityShape {
	return game.ActivatedAbilityShape{
		Label:        a.Label,
		Cost:         a.Cost,
		Targets:      a.Targets,
		Modes:        a.Modes,
		SorcerySpeed: a.SorcerySpeed,
		Purpose:      a.Purpose,
	}
}

// checkManaAbilitiesAnswers checks every mana ability in a list.
func checkManaAbilitiesAnswers(name, where string, in []ManaAbility) {
	for i, m := range in {
		checkManaAnswers(name, fmt.Sprintf("%s mana ability %d", where, i),
			game.ManaAbilityShape{SacrificeOther: m.Cost.SacrificeOther, Answers: m.Answers})
	}
}

// checkSpellAnswers is the guard for what a spell declares it answers
// (#2877, ADR 0142's amendment of 2026-10-30): on the card, a mode or
// an alternative cost, which a cast reads for value_only
// (legal.castAnswersOf). checkPurpose has already checked the bits and
// the slot. This keeps the activated rows' cost coupling (ADR 0142 §3)
// true for spells: a spell whose printed additional cost sacrifices or
// exiles a creature you control (Altar's Reap, Deadly Dispute) is a
// sacrifice outlet however it is cast, so no slot of it may declare
// `value`; nor may an alternative cost that sacrifices or returns a
// creature instead of paying. A cost the reader can see a creature
// given up in at enumeration (a choice, a kicker) it reads there too.
func checkSpellAnswers(spec Spec) {
	fail := func(where, why string) {
		panic(fmt.Sprintf("effects.Register: %q %s declares Answers that %s (ADR 0142's amendment of 2026-10-30; docs/adding-cards.md, \"Declaring what an ability answers\")", spec.Name, where, why))
	}
	const why = "say value on a spell whose cost sacrifices or exiles a creature — that cost is a sacrifice outlet; declare AnswerSacOutlet"
	if legal.AdditionalCostAnswers(spec.AdditionalCost).Has(game.AnswerSacOutlet) {
		if spec.Purpose.Answers == game.AnswerValue {
			fail("card", why)
		}
		if spec.Modes != nil {
			for i, o := range spec.Modes.Options {
				if o.Purpose.Answers == game.AnswerValue {
					fail(fmt.Sprintf("spell mode %d", i), why)
				}
			}
		}
		for _, a := range spec.AlternativeCosts {
			if a.Purpose.Answers == game.AnswerValue {
				fail(fmt.Sprintf("alternative cost %q", a.Key), why)
			}
		}
	}
	for i := range spec.AlternativeCosts {
		a := &spec.AlternativeCosts[i]
		if a.Purpose.Answers == game.AnswerValue && legal.AlternativeCostAnswers(a).Has(game.AnswerSacOutlet) {
			fail(fmt.Sprintf("alternative cost %q", a.Key), "say value on a cost that sacrifices or returns a creature instead of paying — that cost is a sacrifice outlet; declare AnswerSacOutlet")
		}
	}
}
