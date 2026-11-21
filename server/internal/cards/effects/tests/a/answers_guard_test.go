package testsa

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestAnswersGuardOnSlots(t *testing.T) {
	for _, c := range []struct {
		name string
		slot effects.PurposeSlot
		p    game.Purpose
		want string
	}{
		{"activated row", effects.PurposeOnActivated, game.Purpose{Answers: game.AnswerProtect}, ""},
		{"any-player row", effects.PurposeOnAnyPlayerActivated, game.Purpose{Answers: game.AnswerValue}, ""},
		{"pump beside its amounts", effects.PurposeOnActivated, game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}}, ""},
		// #2877, ADR 0142's amendment of 2026-10-30: a spell's slots
		// declare too, for a cast's value_only.
		{"on the card", effects.PurposeOnCard, game.Purpose{Answers: game.AnswerValue}, ""},
		{"on a spell's mode", effects.PurposeOnSpellMode, game.Purpose{Answers: game.AnswerValue}, ""},
		{"on an alternative cost", effects.PurposeOnAltCost, game.Purpose{Answers: game.AnswerRemove}, ""},
		{"on an ability's mode", effects.PurposeOnMode, game.Purpose{Answers: game.AnswerRemove}, "off an activated row"},
		{"on a turn face up", effects.PurposeOnFaceUp, game.Purpose{Answers: game.AnswerValue}, "off an activated row"},
		{"value beside another on the card", effects.PurposeOnCard, game.Purpose{Answers: game.AnswerRemove | game.AnswerValue}, "value stands alone"},
		{"on a triggered row", effects.PurposeOnTriggered, game.Purpose{Answers: game.AnswerPump}, "off an activated row"},
		{"value beside another", effects.PurposeOnActivated, game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerValue}, "value stands alone"},
		{"unknown bit", effects.PurposeOnActivated, game.Purpose{Answers: 1 << 14}, "name no answer"},
	} {
		t.Run(c.name, func(t *testing.T) {
			msg := effectstest.Refusal(func() { effects.CheckPurpose("Test Card", "ability 0", c.slot, c.p) })
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}

func TestAnswersGuardOnActivatedRows(t *testing.T) {
	loyalty := 1
	target := &game.TargetSpec{Label: "target creature", Min: 1, Max: 1}
	upTo := &game.TargetSpec{Label: "up to one target creature", Min: 0, Max: 1}
	for _, c := range []struct {
		name string
		ab   game.ActivatedAbilityShape
		want string
	}{
		{"untargeted instant-speed row", game.ActivatedAbilityShape{Label: "{1}{G}: Regenerate this creature."}, ""},
		{"sorcery speed", game.ActivatedAbilityShape{SorcerySpeed: true}, "sorcery-speed"},
		{"loyalty", game.ActivatedAbilityShape{Cost: game.AbilityCost{Loyalty: &loyalty}}, "sorcery-speed or loyalty"},
		// #4734 slice 7 (ADR 0009's amendment, Q2 (a)): a row that must
		// name a target may declare what it answers.
		{"targeted", game.ActivatedAbilityShape{Targets: target}, ""},
		{"every mode targeted", game.ActivatedAbilityShape{Modes: &game.ModeSpec{Options: []game.ModeOption{{Targets: target}, {Targets: target}}}}, ""},
		{"up to one target", game.ActivatedAbilityShape{Targets: upTo}, ""},
		{"X targets", game.ActivatedAbilityShape{Targets: &game.TargetSpec{Label: "X target creatures", CountFromX: true}}, ""},
		{"one up-to mode", game.ActivatedAbilityShape{Modes: &game.ModeSpec{Options: []game.ModeOption{{Targets: target}, {Targets: upTo}}}}, ""},
		{"one untargeted mode", game.ActivatedAbilityShape{Modes: &game.ModeSpec{Options: []game.ModeOption{{Targets: target}, {}}}}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			c.ab.Purpose.Answers = game.AnswerProtect
			msg := effectstest.Refusal(func() { effects.CheckActivatedAnswers("Test Card", "ability 0", c.ab, false) })
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}

// #4734 slice 7: a row that must name a target may declare any answer
// but `value`, which no reader would act on: its move carries
// has_targets, and a value row gets no answers_now. An "up to one" row
// still declares value, since its empty announcement reads it.
func TestAnswersGuardOnTargetedRows(t *testing.T) {
	target := &game.TargetSpec{Label: "target creature", Min: 1, Max: 1}
	upTo := &game.TargetSpec{Label: "up to one target creature", Min: 0, Max: 1}
	for _, c := range []struct {
		name    string
		targets *game.TargetSpec
		answers game.Answers
		want    string
	}{
		{"targeted protect", target, game.AnswerProtect, ""},
		{"targeted survive and pump", target, game.AnswerSurvive | game.AnswerPump, ""},
		{"targeted remove", target, game.AnswerRemove, ""},
		{"targeted value", target, game.AnswerValue, "say value on a row whose every announcement has a target"},
		{"up to one value", upTo, game.AnswerValue, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			ab := game.ActivatedAbilityShape{Targets: c.targets, Purpose: game.Purpose{Answers: c.answers}}
			msg := effectstest.Refusal(func() { effects.CheckActivatedAnswers("Test Card", "ability 0", ab, false) })
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}

// The cost coupling (ADR 0142 §3): a declaration on a row whose cost is
// a creature outlet keeps the outlet, and a crew row says it animates.
func TestAnswersGuardKeepsTheCostRule(t *testing.T) {
	outlet := effects.SacrificeACreature()
	food := game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))}
	for _, c := range []struct {
		name    string
		cost    game.AbilityCost
		answers game.Answers
		want    string
	}{
		{"outlet declared value", outlet, game.AnswerValue, "sacrifices, exiles or returns a creature"},
		{"outlet declared pump", outlet, game.AnswerPump, "sacrifices, exiles or returns a creature"},
		{"outlet declared sac_outlet", outlet, game.AnswerSacOutlet | game.AnswerPump, ""},
		{"outlet declared protect", outlet, game.AnswerProtect, ""},
		{"outlet declared survive", outlet, game.AnswerSurvive, ""},
		{"food is a price", food, game.AnswerValue, ""},
		{"crew without animate", game.AbilityCost{Crew: 1}, game.AnswerCombatGrant, "crew row"},
		{"crew with animate", game.AbilityCost{Crew: 1}, game.AnswerAnimate, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			ab := game.ActivatedAbilityShape{Cost: c.cost, Purpose: game.Purpose{Answers: c.answers}}
			msg := effectstest.Refusal(func() { effects.CheckActivatedAnswers("Test Card", "ability 0", ab, false) })
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}

func TestAnswersGuardOnManaAbilities(t *testing.T) {
	outlet := effects.SacrificeACreature().SacrificeOther
	for _, c := range []struct {
		name    string
		sac     *game.TargetSpec
		answers game.Answers
		want    string
	}{
		{"altar declared outlet", outlet, game.AnswerSacOutlet, ""},
		{"altar declared value", outlet, game.AnswerValue, "leave out AnswerSacOutlet"},
		{"rock declared value", nil, game.AnswerValue, ""},
		{"rock declared pump", nil, game.AnswerPump, "neither AnswerSacOutlet nor AnswerValue"},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := game.ManaAbilityShape{SacrificeOther: c.sac, Answers: c.answers}
			msg := effectstest.Refusal(func() { effects.CheckManaAnswers("Test Card", "mana ability 0", m) })
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}

// An opponents-only row may declare what it answers, and nothing else
// (ADR 0106's rule, widened by ADR 0142).
func TestAnswersAllowedOnAnOpponentsOnlyRow(t *testing.T) {
	row := effects.ActivatedAbility{OpponentsOnly: true, Purpose: game.Purpose{Answers: game.AnswerValue}}
	if msg := effectstest.Refusal(func() { effects.CheckAnyPlayerAbility("Test Card", "ability 0", row) }); msg != "" {
		t.Fatalf("refused: %s", msg)
	}
	row.Purpose.Draws = 1
	if msg := effectstest.Refusal(func() { effects.CheckAnyPlayerAbility("Test Card", "ability 0", row) }); !strings.Contains(msg, "opponents-only") {
		t.Fatalf("got %q, want the opponents-only refusal", msg)
	}
}

// #2877: a spell keeps the cost coupling (ADR 0142 §3). A cost that
// gives up a creature is a sacrifice outlet, so no slot of the spell may
// say it answers nothing; a land or a discard is a price.
func TestAnswersGuardOnSpellCosts(t *testing.T) {
	creature := effects.SacrificeCost("a creature", effects.Creature())
	land := effects.SacrificeCost("a land", effects.Land())
	value := game.Purpose{Answers: game.AnswerValue}
	outlet := game.Purpose{Answers: game.AnswerSacOutlet}
	creatureInstead := game.AlternativeCost{Key: "test", Label: "sacrifice a creature", Sacrifice: &game.TargetSpec{Label: "a creature", Min: 1, Max: 1}, Purpose: value}
	for _, c := range []struct {
		name string
		spec effects.Spec
		want string
	}{
		{"a cantrip", effects.Spec{Purpose: value}, ""},
		{"a creature sacrifice declared value", effects.Spec{Purpose: value, AdditionalCost: creature}, "sacrifice outlet"},
		{"a creature sacrifice declared sac_outlet", effects.Spec{Purpose: outlet, AdditionalCost: creature}, ""},
		{"a land sacrifice declared value", effects.Spec{Purpose: value, AdditionalCost: land}, ""},
		{"a mode declared value under a creature sacrifice", effects.Spec{AdditionalCost: creature,
			Modes: &game.ModeSpec{Options: []game.ModeOption{{Purpose: value}}}}, "spell mode 0"},
		{"an alternative cost declared value under a creature sacrifice", effects.Spec{AdditionalCost: creature,
			AlternativeCosts: []game.AlternativeCost{{Key: "flashback", Purpose: value}}}, "alternative cost"},
		{"a creature sacrificed instead of paying, declared value", effects.Spec{AlternativeCosts: []game.AlternativeCost{creatureInstead}}, "instead of paying"},
	} {
		t.Run(c.name, func(t *testing.T) {
			c.spec.Name = "Test Card"
			msg := effectstest.Refusal(func() { effects.CheckSpellAnswers(c.spec) })
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}

// #2877: what a kicked cast answers is the card's union, so a kicked
// purpose may not declare Answers of its own.
func TestAnswersRefusedInsideKicked(t *testing.T) {
	spec := effects.Spec{
		Name:          "Test Card",
		OptionalCosts: []game.AdditionalCost{{Key: game.KickerKey, Optional: true, ManaCost: "{1}"}},
		Purpose:       game.Purpose{Answers: game.AnswerValue, Kicked: &game.Purpose{Draws: 1, Answers: game.AnswerValue}},
	}
	if msg := effectstest.Refusal(func() { effects.CheckConditionalPurposes(spec) }); !strings.Contains(msg, "declares Answers") {
		t.Fatalf("got %q, want the kicked Answers refusal", msg)
	}
}
