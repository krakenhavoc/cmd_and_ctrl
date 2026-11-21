package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_lands_test.go — "as this land enters, you may reveal an
// Island or Swamp card from your hand. If you don't, it enters
// tapped."
//
// The assertions that earn their keep are shocklands_test.go's, for
// its reason: the tap EVENT count, not the Tapped flag, is what tells
// a replaced entry from an enter-tapped-then-untap patch-up, and the
// prompt pausing the entry — nothing on the battlefield, the card
// still in hand — is what tells a CR 614 replacement from an ETB
// trigger. Two more belong to this seam alone: the revealed card is
// still in the revealer's HAND afterwards (a reveal is not a zone
// change, CR 701.20b), and the table can read it (CR 701.20).

// revealLandOracleIDs is the whole cycle, name → oracle ID, verified
// against Scryfall.
var RevealLandOracleIDs = map[string]string{
	"Choked Estuary":     "d473b507-8c33-4118-bc10-b0a268776074",
	"Foreboding Ruins":   "5c87e2fa-77f1-4978-b25f-f14d227301d1",
	"Port Town":          "458d2b12-f578-4392-98d3-c3bc83f316c4",
	"Fortified Village":  "56f1a16a-9f41-41fb-b580-c200bca27cd6",
	"Frostboil Snarl":    "7137aae6-260d-41de-8b4e-42a8cf752697",
	"Furycalm Snarl":     "651dea9c-2375-4e44-8e65-ba8e40f0c0ef",
	"Game Trail":         "00de57d2-7cb6-4337-9bc6-f6711e4dfabf",
	"Shineshadow Snarl":  "c9fc13d6-bd10-47bc-b2b6-7f67a1f3371e",
	"Necroblossom Snarl": "761ee6f9-b0fa-43c9-8d1f-9591ea18e52d",
	"Vineglimmer Snarl":  "33f52df8-4b44-4422-8b0a-37fead9c894b",
}

// entryRevealChoiceFor returns the open "you may reveal …" prompt
// addressed to chooser, or nil.
func EntryRevealChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceEntryRevealFromHand && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// answerEntryReveal answers the open entry prompt for chooser. An
// empty `picks` is the decline.
func AnswerEntryReveal(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := EntryRevealChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no PendingChoiceEntryRevealFromHand addressed to %s", chooser)
	}
	if err := g.ResolveEntryRevealFromHand(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveEntryRevealFromHand: %v", err)
	}
}

// seedRevealLandHand puts one card of each named type line into the
// active seat's hand and returns the IDs in the order given.
func SeedRevealLandHand(p *game.Player, typeLines ...string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(typeLines))
	for i, tl := range typeLines {
		out = append(out, HandCard(p, "fixture "+string(rune('a'+i)), tl))
	}
	return out
}
