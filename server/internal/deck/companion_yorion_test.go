package deck

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// companion_yorion_test.go — #4128: Yorion, Sky Nomad's companion
// condition ("at least twenty cards more than the minimum deck size")
// can never be met in Commander, where the minimum is also the maximum
// (CR 903.5a), so the deck check refuses it with its own message
// rather than as an unsupported companion (ADR 0172 §9).

const yorionOracle = "e15504a7-2b67-4185-b916-172145f10b19"

func yorion() cards.Card {
	c := basicLegal("Yorion, Sky Nomad", "Legendary Creature — Bird Serpent", "W", "U")
	c.OracleID = uuid.MustParse(yorionOracle)
	c.ManaCost = "{3}{W/U}{W/U}"
	c.Keywords = []string{"Flying", "Companion"}
	c.OracleText = "Companion — Your starting deck contains at least twenty cards more than the minimum deck size. (If this card is your chosen companion, you may put it into your hand from outside the game for {3} as a sorcery.)\nFlying\nWhen Yorion enters, exile any number of other nonland permanents you own and control. Return those cards to the battlefield at the beginning of the next end step."
	return c
}

func TestYorionCompanionConditionIsNeverMetInCommander(t *testing.T) {
	list := companionDeck()
	y := yorion()
	list.Companion = &y
	err := Validate(list)
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Validate = %v, want a validation error", err)
	}
	var unmet []Violation
	for _, v := range ve.Violations {
		switch v.Code {
		case CodeCompanionConditionNotMet:
			unmet = append(unmet, v)
		case CodeUnsupportedMechanic:
			t.Errorf("Yorion is refused as unsupported (%s), want its condition judged", v.Message)
		}
	}
	if len(unmet) != 1 {
		t.Fatalf("condition violations = %+v, want exactly one", unmet)
	}
	if unmet[0].Card != y.Name {
		t.Errorf("the violation names %q, want the companion %q", unmet[0].Card, y.Name)
	}
	if !strings.Contains(unmet[0].Message, "120") || !strings.Contains(unmet[0].Message, "exactly 100") {
		t.Errorf("message %q does not say why", unmet[0].Message)
	}
}

// The deck-wide form judges the count, not a fixed answer: a starting
// deck twenty over the minimum would pass.
func TestYorionUnmetCountsTheStartingDeck(t *testing.T) {
	if why := yorionUnmet(make([]cards.Card, commanderDeckSize)); why == "" {
		t.Error("a 100-card starting deck meets Yorion's condition")
	}
	if why := yorionUnmet(make([]cards.Card, commanderDeckSize+yorionExtraCards)); why != "" {
		t.Errorf("a 120-card starting deck fails: %s", why)
	}
}

// Every registry entry sets exactly one of its two forms.
func TestCompanionConditionsSetExactlyOneForm(t *testing.T) {
	for oracle, cond := range companionConditions {
		if (cond.Offenders == nil) == (cond.Unmet == nil) {
			t.Errorf("%s: Offenders set %v, Unmet set %v, want exactly one", oracle, cond.Offenders != nil, cond.Unmet != nil)
		}
	}
}
