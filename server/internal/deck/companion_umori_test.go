package deck

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func umoriCard() cards.Card {
	c := basicLegal("Umori, the Collector", "Legendary Creature — Ooze", "B", "G")
	c.OracleID = uuid.MustParse(umoriOracleID)
	c.OracleText = "Companion — Each nonland card in your starting deck shares a card type. (If this card is your chosen companion, you may put it into your hand from outside the game for {3} as a sorcery.)\nAs Umori enters, choose a card type.\nSpells you cast of the chosen type cost {1} less to cast."
	return c
}

func umoriDeck(main ...cards.Card) *List {
	cmd := basicLegal("Five Colour Commander", "Legendary Creature — Dragon", "W", "U", "B", "R", "G")
	list := &List{Commanders: []cards.Card{cmd}}
	for i := range 99 - len(main) {
		_ = i
		list.Mainboard = append(list.Mainboard, basicLegal("Filler", "Artifact Creature — Golem"))
	}
	list.Mainboard = append(list.Mainboard, main...)
	u := umoriCard()
	list.Companion = &u
	return list
}

func umoriViolations(t *testing.T, list *List) []Violation {
	t.Helper()
	err := Validate(list, game.FormatCommander)
	ve, ok := err.(*ValidationError)
	if !ok {
		return nil
	}
	var out []Violation
	for _, v := range ve.Violations {
		if v.Code == CodeUnsupportedMechanic {
			t.Fatalf("Umori refused as unsupported: %s", v.Message)
		}
		if v.Code == CodeCompanionConditionNotMet {
			out = append(out, v)
		}
	}
	return out
}

func TestUmoriConditionPassesWhenAllNonlandsShareACreatureTypeAndLandsAreIgnored(t *testing.T) {
	list := umoriDeck(
		basicLegal("Forest", "Basic Land — Forest"),
		basicLegal("Bear", "Creature — Bear"),
		basicLegal("Enchanted Elk", "Enchantment Creature — Elk"),
	)
	if vs := umoriViolations(t, list); len(vs) != 0 {
		t.Fatalf("violations %+v, want none", vs)
	}
}

func TestUmoriConditionFailsWhenNoTypeIsShared(t *testing.T) {
	list := umoriDeck(basicLegal("Shock", "Instant"))
	// Fillers are artifact creatures; an Instant shares nothing with them.
	vs := umoriViolations(t, list)
	if len(vs) != 1 {
		t.Fatalf("violations %+v, want exactly one", vs)
	}
	if vs[0].Card != "Umori, the Collector" || !strings.Contains(vs[0].Message, "Shock") {
		t.Errorf("violation %+v should name Umori and the card to cut", vs[0])
	}
}

func TestUmoriCompanionCardTypes(t *testing.T) {
	split := basicLegal("Fire // Ice", "Instant // Instant")
	split.Layout = "split"
	split.CardFaces = []cards.CardFace{{TypeLine: "Instant"}, {TypeLine: "Sorcery"}}
	if got := companionCardTypes(split); !got["Instant"] || !got["Sorcery"] {
		t.Errorf("a split card has both halves' types, got %v", got)
	}
	mdfc := basicLegal("Front // Back", "Creature — Elf // Land")
	mdfc.Layout = "modal_dfc"
	if got := companionCardTypes(mdfc); !got["Creature"] || got["Land"] {
		t.Errorf("a modal DFC outside the game has only its front face's types, got %v", got)
	}
	tribal := basicLegal("Changeling Outcast", "Kindred Instant — Goblin")
	if got := companionCardTypes(tribal); !got["Kindred"] || !got["Instant"] || got["Goblin"] {
		t.Errorf("types of a kindred instant: %v", got)
	}
}
