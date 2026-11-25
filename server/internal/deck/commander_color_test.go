package deck

import (
	"fmt"
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prismaticPiper is The Prismatic Piper as the Scryfall record has it:
// colourless, with CR 903.4b's ability and partner.
func prismaticPiper() cards.Card {
	c := costed("The Prismatic Piper", "Legendary Creature — Shapeshifter", "{5}")
	c.Keywords = []string{"Partner"}
	c.OracleText = "If The Prismatic Piper is your commander, choose a color before the game begins. The Prismatic Piper is the chosen color.\nPartner (You can have two commanders if both have partner.)"
	return c
}

// piperDeck is a 100-card deck led by the Piper whose 99 have the given
// colour identities, in turn.
func piperDeck(identities ...[]string) *List {
	list := &List{Commanders: []cards.Card{prismaticPiper()}}
	for i := range 99 {
		var id []string
		if len(identities) > 0 {
			id = identities[i%len(identities)]
		}
		list.Mainboard = append(list.Mainboard, costed(fmt.Sprintf("Card %d", i), "Artifact", "{2}", id...))
	}
	return list
}

func TestChoosesColorBeforeGameReadsThePrintedSentence(t *testing.T) {
	if !choosesColorBeforeGame(prismaticPiper()) {
		t.Error("The Prismatic Piper: the CR 903.4b ability was not read")
	}
	if choosesColorBeforeGame(costed("Grizzly Bears", "Creature — Bear", "{1}{G}", "G")) {
		t.Error("Grizzly Bears read as choosing a colour")
	}
}

// CR 903.4b: the choice applies during deck construction, so a Piper deck
// may hold cards of one colour, the colour it chose.
func TestPiperDeckOfOneColourValidates(t *testing.T) {
	if err := Validate(piperDeck([]string{"R"}, nil), game.FormatCommander); err != nil {
		t.Fatalf("a mono-red Piper deck: %v", err)
	}
	if err := Validate(piperDeck(nil), game.FormatCommander); err != nil {
		t.Fatalf("a colourless Piper deck: %v", err)
	}
}

// One choice is one colour: a Piper deck of two colours is not legal.
func TestPiperDeckOfTwoColoursIsRefused(t *testing.T) {
	codes := violationCodes(Validate(piperDeck([]string{"R"}, []string{"G"}), game.FormatCommander))
	if !slices.Contains(codes, CodeColorIdentity) {
		t.Fatalf("codes %v, want %s", codes, CodeColorIdentity)
	}
}

// With a partner, the chosen colour adds to the partner's identity
// (CR 702.124c): a blue partner and a Piper make a two-colour deck, not
// a three-colour one.
func TestPiperWithAPartnerAddsOneColour(t *testing.T) {
	partner := costed("Blue Partner", "Legendary Creature — Human", "{1}{U}", "U")
	partner.Keywords = []string{"Partner"}
	partner.OracleText = "Partner (You can have two commanders if both have partner.)"

	list := piperDeck([]string{"U"}, []string{"B"})
	list.Commanders = append(list.Commanders, partner)
	list.Mainboard = list.Mainboard[:98]
	if err := Validate(list, game.FormatCommander); err != nil {
		t.Fatalf("Piper (black) + blue partner, a blue-black deck: %v", err)
	}

	list = piperDeck([]string{"U"}, []string{"B"}, []string{"R"})
	list.Commanders = append(list.Commanders, partner)
	list.Mainboard = list.Mainboard[:98]
	codes := violationCodes(Validate(list, game.FormatCommander))
	if !slices.Contains(codes, CodeColorIdentity) {
		t.Fatalf("Piper + blue partner, a three-colour deck: codes %v, want %s", codes, CodeColorIdentity)
	}
}

// The companion must fit the identity the choice makes, and counts
// toward it.
func TestPiperCompanionFitsTheChosenColour(t *testing.T) {
	list := piperDeck([]string{"R"})
	comp := costed("Red Companion", "Creature — Elemental", "{2}{R}", "R")
	list.Companion = &comp
	for _, v := range violations(Validate(list, game.FormatCommander)) {
		if v.Code == CodeCompanionColorIdentity || v.Code == CodeColorIdentity {
			t.Fatalf("a red companion in a mono-red Piper deck: %+v", v)
		}
	}
	comp = costed("Green Companion", "Creature — Elemental", "{2}{G}", "G")
	list.Companion = &comp
	codes := violationCodes(Validate(list, game.FormatCommander))
	if !slices.Contains(codes, CodeColorIdentity) && !slices.Contains(codes, CodeCompanionColorIdentity) {
		t.Fatalf("a green companion in a red Piper deck: codes %v, want an identity violation", codes)
	}
}

func violations(err error) []Violation {
	ve, ok := err.(*ValidationError)
	if !ok {
		return nil
	}
	return ve.Violations
}
