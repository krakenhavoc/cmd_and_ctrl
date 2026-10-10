package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answers_guard_test.go — ADR 0142 decision 1: what effects.Register
// refuses in a declared Purpose.Answers or ManaAbility.Answers.

// refusal runs fn and returns its panic message, "" for none.
func refusal(fn func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg, _ = r.(string)
		}
	}()
	fn()
	return ""
}

func TestAnswersGuardOnSlots(t *testing.T) {
	for _, c := range []struct {
		name string
		slot purposeSlot
		p    game.Purpose
		want string
	}{
		{"activated row", purposeOnActivated, game.Purpose{Answers: game.AnswerProtect}, ""},
		{"any-player row", purposeOnAnyPlayerActivated, game.Purpose{Answers: game.AnswerValue}, ""},
		{"pump beside its amounts", purposeOnActivated, game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}}, ""},
		{"on the card", purposeOnCard, game.Purpose{Answers: game.AnswerValue}, "off an activated row"},
		{"on a mode", purposeOnMode, game.Purpose{Answers: game.AnswerRemove}, "off an activated row"},
		{"on an alternative cost", purposeOnAltCost, game.Purpose{Answers: game.AnswerRemove}, "off an activated row"},
		{"on a triggered row", purposeOnTriggered, game.Purpose{Answers: game.AnswerPump}, "off an activated row"},
		{"value beside another", purposeOnActivated, game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerValue}, "value stands alone"},
		{"unknown bit", purposeOnActivated, game.Purpose{Answers: 1 << 14}, "name no answer"},
	} {
		t.Run(c.name, func(t *testing.T) {
			msg := refusal(func() { checkPurpose("Test Card", "ability 0", c.slot, c.p) })
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
		{"targeted", game.ActivatedAbilityShape{Targets: target}, "every announcement has a target"},
		{"every mode targeted", game.ActivatedAbilityShape{Modes: &game.ModeSpec{Options: []game.ModeOption{{Targets: target}, {Targets: target}}}}, "every announcement has a target"},
		{"up to one target", game.ActivatedAbilityShape{Targets: upTo}, ""},
		{"X targets", game.ActivatedAbilityShape{Targets: &game.TargetSpec{Label: "X target creatures", CountFromX: true}}, "every announcement has a target"},
		{"one up-to mode", game.ActivatedAbilityShape{Modes: &game.ModeSpec{Options: []game.ModeOption{{Targets: target}, {Targets: upTo}}}}, ""},
		{"one untargeted mode", game.ActivatedAbilityShape{Modes: &game.ModeSpec{Options: []game.ModeOption{{Targets: target}, {}}}}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			c.ab.Purpose.Answers = game.AnswerProtect
			msg := refusal(func() { checkActivatedAnswers("Test Card", "ability 0", c.ab) })
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
	outlet := SacrificeACreature()
	food := game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))}
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
		{"food is a price", food, game.AnswerValue, ""},
		{"crew without animate", game.AbilityCost{Crew: 1}, game.AnswerCombatGrant, "crew row"},
		{"crew with animate", game.AbilityCost{Crew: 1}, game.AnswerAnimate, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			ab := game.ActivatedAbilityShape{Cost: c.cost, Purpose: game.Purpose{Answers: c.answers}}
			msg := refusal(func() { checkActivatedAnswers("Test Card", "ability 0", ab) })
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
	outlet := SacrificeACreature().SacrificeOther
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
			msg := refusal(func() { checkManaAnswers("Test Card", "mana ability 0", m) })
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
	row := ActivatedAbility{OpponentsOnly: true, Purpose: game.Purpose{Answers: game.AnswerValue}}
	if msg := refusal(func() { checkAnyPlayerAbility("Test Card", "ability 0", row) }); msg != "" {
		t.Fatalf("refused: %s", msg)
	}
	row.Purpose.Draws = 1
	if msg := refusal(func() { checkAnyPlayerAbility("Test Card", "ability 0", row) }); !strings.Contains(msg, "opponents-only") {
		t.Fatalf("got %q, want the opponents-only refusal", msg)
	}
}
