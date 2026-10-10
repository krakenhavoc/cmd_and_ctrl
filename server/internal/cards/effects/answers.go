package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// answers.go — ADR 0142: the registration guard for a declared
// Purpose.Answers on an activated row, and ManaAbility.Answers on a mana
// ability. The vocabulary and its tiers are game/answers.go; the reader
// is internal/legal's untargetedFlags (manaMoveInteracts for a mana
// ability).
//
// A declaration is the only thing the reader reads, so the guard keeps
// it honest where it can without reading text. It refuses:
//
//   - a bit that names no answer;
//   - AnswerValue beside any other answer;
//   - Answers on a spell, a mode, an alternative cost or a triggered row,
//     which nothing reads it on yet (checkPurpose, ADR 0142 §2);
//   - Answers on a row with a target clause and no untargeted mode: its
//     move carries has_targets, which already stops smart autopass;
//   - Answers on a sorcery-speed row, which is never offered while an
//     item waits on the stack;
//   - a crew row (AbilityCost.Crew) that does not declare AnswerAnimate;
//   - a row whose cost sacrifices, exiles or returns a creature
//     (legal.CostAnswers says sac_outlet) that declares neither
//     AnswerSacOutlet nor AnswerProtect, so the #2853 cost rule stays true
//     for every declared row (ADR 0142 §3);
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
// template's.
func checkActivatedAnswers(name, where string, ab game.ActivatedAbilityShape) {
	a := ab.Purpose.Answers
	if a == 0 {
		return
	}
	fail := func(why string) {
		panic(fmt.Sprintf("effects.Register: %q %s declares Answers that %s (ADR 0142; docs/adding-cards.md, \"Declaring what an ability answers\")", name, where, why))
	}
	checkAnswerBits(a, fail)
	if sorcerySpeedRow(ab) {
		fail("are on a sorcery-speed or loyalty row, which is never offered while an item waits on the stack")
	}
	if !hasUntargetedAnnouncement(ab) {
		fail("are on a row whose every announcement has a target — its move carries has_targets, which already stops smart autopass")
	}
	if ab.Cost.Crew > 0 && !a.Has(game.AnswerAnimate) {
		fail("leave out AnswerAnimate on a crew row — crewing makes the Vehicle a creature")
	}
	if legal.CostAnswers(ab.Cost).Has(game.AnswerSacOutlet) && !a.HasAny(game.AnswerSacOutlet|game.AnswerProtect) {
		fail(fmt.Sprintf("(%s) leave out AnswerSacOutlet and AnswerProtect on a row whose cost sacrifices, exiles or returns a creature", a))
	}
}

// sorcerySpeedRow reports whether the row is only ever activated as a
// sorcery: "activate only as a sorcery" (CR 602.5d), or a loyalty
// ability (CR 606.3).
func sorcerySpeedRow(ab game.ActivatedAbilityShape) bool {
	return ab.SorcerySpeed || ab.Cost.Loyalty != nil
}

// answersInScope reports whether smart autopass reads the row's answers:
// it may be activated at instant speed with no target (ADR 0142 §4).
// TestEveryAnswersRowDeclares fails on such a catalog row that declares
// nothing.
func answersInScope(ab game.ActivatedAbilityShape) bool {
	return !sorcerySpeedRow(ab) && hasUntargetedAnnouncement(ab)
}

// hasUntargetedAnnouncement reports whether the row can be announced
// with no target: it has no target clause and, if it is modal, at least
// one mode with none.
func hasUntargetedAnnouncement(ab game.ActivatedAbilityShape) bool {
	if ab.Targets != nil {
		return false
	}
	if ab.Modes == nil {
		return true
	}
	for _, o := range ab.Modes.Options {
		if o.Targets == nil {
			return true
		}
	}
	return false
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
