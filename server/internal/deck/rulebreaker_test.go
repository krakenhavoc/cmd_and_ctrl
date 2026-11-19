package deck

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func rulebreakerDeck(commander cards.Card, offender cards.Card) *List {
	list := &List{Commanders: []cards.Card{commander}}
	basic := basicLegal("Forest", "Basic Land — Forest", "G")
	for i := 0; i < 99; i++ {
		list.Mainboard = append(list.Mainboard, basic)
	}
	list.Mainboard[0] = offender
	return list
}

func grizzlegomCard() cards.Card {
	c := basicLegal("Grizzlegom, Hurloon Hero", "Legendary Creature — Minotaur Warrior", "R", "G")
	c.OracleText = "Rulebreaker — A deck with this commander can have any land cards.\nWhenever Grizzlegom attacks, create a 1/1 white Soldier creature token for each Plains you control."
	return c
}

func maularCard() cards.Card {
	c := basicLegal("Maular, the Next Evolution", "Legendary Creature — Dinosaur Mutant", "G")
	c.OracleText = "Rulebreaker — A deck with this commander can have creature cards with mana value 7 or greater of any color identity and any basic land cards.\nWhenever a creature you control with mana value 7 or greater attacks, double its power and toughness until end of turn."
	return c
}

func TestRulebreakerAnyLandLiftsTheIdentityOfLandsOnly(t *testing.T) {
	land := basicLegal("Watery Tomb", "Land", "U", "B")
	if err := Validate(rulebreakerDeck(grizzlegomCard(), land), game.FormatCommander); err != nil {
		t.Errorf("an off-identity nonbasic land under Grizzlegom: %v", err)
	}
	spell := basicLegal("Counterspell", "Instant", "U")
	spell.ManaCost = "{U}{U}"
	assertHasViolation(t, Validate(rulebreakerDeck(grizzlegomCard(), spell), game.FormatCommander), CodeColorIdentity)
}

func TestRulebreakerWithoutTheClauseStaysStrict(t *testing.T) {
	land := basicLegal("Watery Tomb", "Land", "U", "B")
	plain := basicLegal("Plain Commander", "Legendary Creature — Human", "G")
	assertHasViolation(t, Validate(rulebreakerDeck(plain, land), game.FormatCommander), CodeColorIdentity)
}

func TestRulebreakerMaularLiftsBasicLandsAndBigCreaturesOnly(t *testing.T) {
	plains := basicLegal("Plains", "Basic Land — Plains", "W")
	if err := Validate(rulebreakerDeck(maularCard(), plains), game.FormatCommander); err != nil {
		t.Errorf("a Plains under Maular: %v", err)
	}
	big := basicLegal("Big Blue Beast", "Creature — Beast", "U")
	big.ManaCost = "{5}{U}{U}"
	if err := Validate(rulebreakerDeck(maularCard(), big), game.FormatCommander); err != nil {
		t.Errorf("a mana value 7 blue creature under Maular: %v", err)
	}
	small := basicLegal("Small Blue Beast", "Creature — Beast", "U")
	small.ManaCost = "{4}{U}{U}"
	assertHasViolation(t, Validate(rulebreakerDeck(maularCard(), small), game.FormatCommander), CodeColorIdentity)
	nonbasic := basicLegal("Watery Tomb", "Land", "U", "B")
	assertHasViolation(t, Validate(rulebreakerDeck(maularCard(), nonbasic), game.FormatCommander), CodeColorIdentity)
	bigSpell := basicLegal("Big Blue Sorcery", "Sorcery", "U")
	bigSpell.ManaCost = "{6}{U}"
	assertHasViolation(t, Validate(rulebreakerDeck(maularCard(), bigSpell), game.FormatCommander), CodeColorIdentity)
}
