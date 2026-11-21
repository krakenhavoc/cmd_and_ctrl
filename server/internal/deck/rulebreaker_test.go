package deck

import (
	"errors"
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

// GAP batch 1: Valko Indorian and The Everforger.
func rulebreakerSwampDeck(commander cards.Card, extra ...cards.Card) *List {
	list := &List{Commanders: []cards.Card{commander}}
	for i := 0; i < 99; i++ {
		list.Mainboard = append(list.Mainboard, basicLegal("Swamp", "Basic Land — Swamp", "B"))
	}
	copy(list.Mainboard, extra)
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
	list := rulebreakerSwampDeck(cmd,
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
	list := rulebreakerSwampDeck(cmd,
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
	list := rulebreakerSwampDeck(cmd, basicLegal("Mountain", "Basic Land — Mountain", "R"))
	if err := Validate(list, game.FormatCommander); !hasColorIdentityViolation(err, "Mountain") {
		t.Errorf("a Mountain under a black commander without the clause was allowed: %v", err)
	}
}

// GAP batch 3: Grizzlegom and Maular.
func rulebreakerForestDeck(commander cards.Card, offender cards.Card) *List {
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
	if err := Validate(rulebreakerForestDeck(grizzlegomCard(), land), game.FormatCommander); err != nil {
		t.Errorf("an off-identity nonbasic land under Grizzlegom: %v", err)
	}
	spell := basicLegal("Counterspell", "Instant", "U")
	spell.ManaCost = "{U}{U}"
	assertHasViolation(t, Validate(rulebreakerForestDeck(grizzlegomCard(), spell), game.FormatCommander), CodeColorIdentity)
}

func TestRulebreakerWithoutTheClauseStaysStrict(t *testing.T) {
	land := basicLegal("Watery Tomb", "Land", "U", "B")
	plain := basicLegal("Plain Commander", "Legendary Creature — Human", "G")
	assertHasViolation(t, Validate(rulebreakerForestDeck(plain, land), game.FormatCommander), CodeColorIdentity)
}

func TestRulebreakerMaularLiftsBasicLandsAndBigCreaturesOnly(t *testing.T) {
	plains := basicLegal("Plains", "Basic Land — Plains", "W")
	if err := Validate(rulebreakerForestDeck(maularCard(), plains), game.FormatCommander); err != nil {
		t.Errorf("a Plains under Maular: %v", err)
	}
	big := basicLegal("Big Blue Beast", "Creature — Beast", "U")
	big.ManaCost = "{5}{U}{U}"
	if err := Validate(rulebreakerForestDeck(maularCard(), big), game.FormatCommander); err != nil {
		t.Errorf("a mana value 7 blue creature under Maular: %v", err)
	}
	small := basicLegal("Small Blue Beast", "Creature — Beast", "U")
	small.ManaCost = "{4}{U}{U}"
	assertHasViolation(t, Validate(rulebreakerForestDeck(maularCard(), small), game.FormatCommander), CodeColorIdentity)
	nonbasic := basicLegal("Watery Tomb", "Land", "U", "B")
	assertHasViolation(t, Validate(rulebreakerForestDeck(maularCard(), nonbasic), game.FormatCommander), CodeColorIdentity)
	bigSpell := basicLegal("Big Blue Sorcery", "Sorcery", "U")
	bigSpell.ManaCost = "{6}{U}"
	assertHasViolation(t, Validate(rulebreakerForestDeck(maularCard(), bigSpell), game.FormatCommander), CodeColorIdentity)
}
