package deck

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const unluckiestText = "Rulebreaker — A deck with this commander can have Aura cards of any color identity and any basic land cards.\n+1: Add one mana of any color."

func gapB06Deck(commander cards.Card, extra ...cards.Card) *List {
	list := &List{Commanders: []cards.Card{commander}}
	list.Mainboard = append(list.Mainboard, extra...)
	for len(list.Mainboard) < 99 {
		list.Mainboard = append(list.Mainboard, basicLegal("Mountain", "Basic Land — Mountain", "R"))
	}
	return list
}

func gapB06Violators(err error) []string {
	ve, ok := err.(*ValidationError)
	if !ok {
		return nil
	}
	var out []string
	for _, v := range ve.Violations {
		if v.Code == CodeColorIdentity {
			out = append(out, v.Card)
		}
	}
	return out
}

// The Unluckiest Planeswalker: Aura cards of any colour identity and any basic
// land, and nothing else.
func TestRulebreakerUnluckiestAllowsAurasAndBasics(t *testing.T) {
	walker := basicLegal("The Unluckiest Planeswalker", "Legendary Planeswalker", "R")
	walker.OracleText = unluckiestText
	aura := basicLegal("Blue Aura", "Enchantment — Aura", "U")
	list := gapB06Deck(walker, aura, basicLegal("Island", "Basic Land — Island", "U"))
	if err := Validate(list, game.FormatCommander); err != nil {
		t.Errorf("an off-colour Aura and basic land: %v", err)
	}
	list = gapB06Deck(walker, aura, basicLegal("Blue Bear", "Creature — Bear", "U"), basicLegal("Blue Bolt", "Instant", "U"))
	if names := gapB06Violators(Validate(list, game.FormatCommander)); len(names) != 2 {
		t.Errorf("violations %v, want the Bear and the Bolt only", names)
	}
	plain := basicLegal("Plain Walker", "Legendary Planeswalker", "R")
	list = gapB06Deck(plain, aura)
	if names := gapB06Violators(Validate(list, game.FormatCommander)); len(names) != 1 {
		t.Errorf("without the clause, violations %v, want the Aura", names)
	}
}
