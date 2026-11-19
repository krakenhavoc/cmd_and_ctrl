package deck

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func rulebreakerDeck(commander cards.Card, extra ...cards.Card) *List {
	list := &List{Commanders: []cards.Card{commander}}
	for i := 0; i < 99; i++ {
		list.Mainboard = append(list.Mainboard, basicLegal("Swamp", "Basic Land — Swamp", "B"))
	}
	for i, c := range extra {
		list.Mainboard[i] = c
	}
	return list
}

func hasColorIdentityViolation(err error, card string) bool {
	var ve *ValidationError
	if !errors.As(err, &ve) {
		return false
	}
	for _, v := range ve.Violations {
		if v.Code == CodeColorIdentity && v.Card == card {
			return true
		}
	}
	return false
}

func TestRulebreakerPhyrexianCommanderLiftsIdentityForPhyrexiansAndBasics(t *testing.T) {
	cmd := basicLegal("Valko Indorian", "Legendary Creature — Human Wizard", "B")
	cmd.OracleText = "Rulebreaker — A deck with this commander can have Phyrexian cards of any color identity and any basic land cards.\nPhyrexian creatures you control have menace and lifelink."
	list := rulebreakerDeck(cmd,
		basicLegal("Phyrexian Rager", "Creature — Phyrexian Horror", "R"),
		basicLegal("Mountain", "Basic Land — Mountain", "R"),
		basicLegal("Lightning Bolt", "Instant", "R"),
		basicLegal("Wurmcoil Engine", "Artifact Creature — Phyrexian Wurm", "R"),
	)
	err := Validate(list, game.FormatCommander)
	if hasColorIdentityViolation(err, "Phyrexian Rager") || hasColorIdentityViolation(err, "Mountain") || hasColorIdentityViolation(err, "Wurmcoil Engine") {
		t.Errorf("a Phyrexian card or a basic land was refused: %v", err)
	}
	if !hasColorIdentityViolation(err, "Lightning Bolt") {
		t.Errorf("an off-identity non-Phyrexian spell was allowed: %v", err)
	}
}

func TestRulebreakerEverforgerLiftsIdentityForArtifactCreaturesAndEquipment(t *testing.T) {
	cmd := basicLegal("The Everforger", "Legendary Artifact Creature — Construct")
	cmd.OracleText = "Rulebreaker — A deck with this commander can have artifact creature and Equipment cards of any color identity and any basic land cards.\nWhenever you cast an artifact creature or Equipment spell, you may copy that spell."
	list := rulebreakerDeck(cmd,
		basicLegal("Goblin Welder", "Creature — Goblin Artificer", "R"),
		basicLegal("Steel Golem", "Artifact Creature — Golem", "R"),
		basicLegal("Sword of Fire and Ice", "Artifact — Equipment", "R", "U"),
		basicLegal("Sol Ring", "Artifact"),
		basicLegal("Island", "Basic Land — Island", "U"),
	)
	err := Validate(list, game.FormatCommander)
	for _, ok := range []string{"Steel Golem", "Sword of Fire and Ice", "Island"} {
		if hasColorIdentityViolation(err, ok) {
			t.Errorf("%s was refused: %v", ok, err)
		}
	}
	if !hasColorIdentityViolation(err, "Goblin Welder") {
		t.Errorf("a red non-artifact creature was allowed: %v", err)
	}
}

func TestRulebreakerDoesNothingWithoutTheClause(t *testing.T) {
	cmd := basicLegal("Plain Commander", "Legendary Creature — Human", "B")
	list := rulebreakerDeck(cmd, basicLegal("Mountain", "Basic Land — Mountain", "R"))
	if err := Validate(list, game.FormatCommander); !hasColorIdentityViolation(err, "Mountain") {
		t.Errorf("a Mountain under a black commander without the clause was allowed: %v", err)
	}
}
