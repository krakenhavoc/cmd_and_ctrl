package deck

import (
	"fmt"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	selumaText  = "Rulebreaker — A deck with this commander can have Angel cards of any color identity and any basic land cards.\nFlying"
	tolabowText = "Rulebreaker — If Tolabow, Loch Rascal is your commander, the color identity of instant and sorcery cards in your deck can include one color of your choice not in your commander's color identity, and your deck can have any basic land cards.\nWhenever you cast an instant or sorcery spell, create a 1/1 blue and red Otter creature token with prowess."
)

func rulebreakerDeck(commander cards.Card, extra ...cards.Card) *List {
	list := &List{Commanders: []cards.Card{commander}}
	list.Mainboard = append(list.Mainboard, extra...)
	for i := 0; len(list.Mainboard) < 99; i++ {
		list.Mainboard = append(list.Mainboard, basicLegal("Plains", "Basic Land — Plains", "W"))
	}
	return list
}

func violationNames(err error) []string {
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

// Seluma: Angels of any color identity and basic lands of any colour; nothing else.
func TestRulebreakerSelumaAllowsAngelsAndBasics(t *testing.T) {
	seluma := basicLegal("Seluma, Light of Aysen", "Legendary Creature — Angel Warrior", "W")
	seluma.OracleText = selumaText
	angel := basicLegal("Red Angel", "Creature — Angel", "R")
	list := rulebreakerDeck(seluma, angel, basicLegal("Mountain", "Basic Land — Mountain", "R"), basicLegal("Swamp", "Basic Land — Swamp", "B"))
	if err := Validate(list, game.FormatCommander); err != nil {
		t.Errorf("a Seluma deck with an off-colour Angel and basics: %v", err)
	}
	list = rulebreakerDeck(seluma, angel, basicLegal("Lightning Bolt", "Instant", "R"), basicLegal("Red Bear", "Creature — Bear", "R"))
	names := violationNames(Validate(list, game.FormatCommander))
	if len(names) != 2 {
		t.Errorf("violations %v, want the Bolt and the Bear only", names)
	}
}

// Without the clause, the same deck is refused.
func TestRulebreakerNeedsTheCommander(t *testing.T) {
	plain := basicLegal("Plain Commander", "Legendary Creature — Angel", "W")
	list := rulebreakerDeck(plain, basicLegal("Red Angel", "Creature — Angel", "R"), basicLegal("Mountain", "Basic Land — Mountain", "R"))
	if names := violationNames(Validate(list, game.FormatCommander)); len(names) != 2 {
		t.Errorf("violations %v, want the Angel and the Mountain", names)
	}
}

// Tolabow: one off-identity colour for instants and sorceries, and any basics.
func TestRulebreakerTolabowAllowsOneSpellColour(t *testing.T) {
	tolabow := basicLegal("Tolabow, Loch Rascal", "Legendary Creature — Otter", "U")
	tolabow.OracleText = tolabowText
	var spells []cards.Card
	for i := 0; i < 5; i++ {
		spells = append(spells, basicLegal(fmt.Sprintf("Red Spell %d", i), "Instant", "R"))
	}
	spells = append(spells, basicLegal("Red Sorcery", "Sorcery", "R"), basicLegal("Mountain", "Basic Land — Mountain", "R"))
	if err := Validate(rulebreakerDeck(tolabow, spells...), game.FormatCommander); err != nil {
		t.Errorf("one off colour for spells: %v", err)
	}
	// A second off colour, a creature, and a two-colour spell are not covered.
	more := append(append([]cards.Card(nil), spells...),
		basicLegal("Black Spell", "Instant", "B"),
		basicLegal("Red Bear", "Creature — Bear", "R"),
		basicLegal("Red Green Spell", "Instant", "R", "G"))
	names := violationNames(Validate(rulebreakerDeck(tolabow, more...), game.FormatCommander))
	if len(names) != 3 {
		t.Errorf("violations %v, want the Black Spell, the Bear and the two-colour spell", names)
	}
}

// Whtz: no maximum deck size, but never fewer than 100 cards.
func TestRulebreakerWhtzHasNoMaximumDeckSize(t *testing.T) {
	whtz := basicLegal("Whtz, the Bibliophile", "Legendary Creature — Homunculus", "W", "U")
	whtz.OracleText = "Rulebreaker — A deck with this commander has no maximum deck size.\n{3}, {T}: You draw a card and gain 1 life."
	list := rulebreakerDeck(whtz)
	for i := 0; i < 150; i++ {
		list.Mainboard = append(list.Mainboard, basicLegal("Plains", "Basic Land — Plains", "W"))
	}
	if err := Validate(list, game.FormatCommander); err != nil {
		t.Errorf("a 250-card Whtz deck: %v", err)
	}
	list.Mainboard = list.Mainboard[:50]
	assertHasViolation(t, Validate(list, game.FormatCommander), CodeWrongCardCount)
	plain := rulebreakerDeck(basicLegal("Plain Commander", "Legendary Creature — Homunculus", "W", "U"))
	plain.Mainboard = append(plain.Mainboard, basicLegal("Plains", "Basic Land — Plains", "W"))
	assertHasViolation(t, Validate(plain, game.FormatCommander), CodeWrongCardCount)
}
