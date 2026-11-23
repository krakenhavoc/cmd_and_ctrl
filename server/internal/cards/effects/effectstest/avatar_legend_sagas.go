package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_legend_sagas_test.go — issue #343: the four Avatar: The Last
// Airbender transforming Sagas, each a Saga front face that at
// chapter III exiles itself and returns transformed as a legendary
// Avatar creature (ADR 0079's second verb). Built on
// transform_cards_test.go's importAndCast / importToBattlefield /
// assertFaceInvariant fixtures for the reason that file gives: a
// transform card's whole point is Faces[1], which only reaches
// game.Card through deck.ToGameCard, so a fixture built with
// castCatalogSpell (no faces at all) would exercise a CanTransform
// that always answers false.
//
// Each card gets the same four checks the task asked for: the Saga
// enters with a lore counter and chapter I fires; chapter III exiles
// and returns the card transformed under its controller's control as
// the Avatar face; the SBA does not sacrifice a Saga that has left
// (folded into the chapter III assertion, exactly as
// TestFableChapterThreeExilesAndReturnsItTransformed does — the
// original object is gone but not in the graveyard); and one
// back-face ability works.

// avatarSagaRow builds the Scryfall-shaped fixture for one of the
// four Sagas. Not transformRow: that helper paints ONE Colors slice
// onto both faces, which would wrongly give the colourless Avatar
// creature the Saga's colour.
func AvatarSagaRow(oracleID, sagaName, sagaCost, avatarName, power, toughness, sagaColor string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     sagaName + " // " + avatarName,
		Layout:   "transform",
		TypeLine: "Enchantment — Saga // Legendary Creature — Avatar",
		OracleID: uuid.MustParse(oracleID),
		CardFaces: []cards.CardFace{
			{Name: sagaName, TypeLine: "Enchantment — Saga", ManaCost: sagaCost, Colors: []string{sagaColor}},
			{Name: avatarName, TypeLine: "Legendary Creature — Avatar", Power: power, Toughness: toughness, Colors: []string{}},
		},
	}
}

func KyoshiRow() cards.Card {
	return AvatarSagaRow(effects.TheLegendOfKyoshiOracleID, "The Legend of Kyoshi", "{4}{G}{G}", "Avatar Kyoshi", "5", "4", "G")
}

func YangchenRow() cards.Card {
	return AvatarSagaRow(effects.TheLegendOfYangchenOracleID, "The Legend of Yangchen", "{3}{W}{W}", "Avatar Yangchen", "4", "5", "W")
}

// --- shared prompt helpers, one per pending-choice kind these cards
// --- raise that no existing test file already answers -------------

// answerScryKeepAll resolves a scry prompt keeping every card on top
// in the order it was shown — Kuruk's "scry 2, then draw" chapters
// don't care about order, only that the prompt is answered so the
// "then" runs.
func AnswerScryKeepAll(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	c := ScryChoiceFor(g, chooser)
	if c == nil {
		t.Fatal("no scry prompt to answer")
	}
	if err := g.ResolveScry(c.ID, chooser, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
}

// resolveManaPickAnyColor answers a "add mana of any color" pipe
// prompt with its first offered colour and returns which one it
// picked.
func ResolveManaPickAnyColor(t *testing.T, g *game.Game, chooser uuid.UUID) string {
	t.Helper()
	c := LatestChoiceOfKind(g, game.PendingChoiceMana)
	if c == nil || c.Chooser != chooser {
		t.Fatal("no mana_pick prompt to answer")
	}
	if len(c.ColorOptions) == 0 {
		t.Fatal("mana_pick prompt has no color options")
	}
	color := c.ColorOptions[0]
	if err := g.ResolveManaChoice(c.ID, chooser, color); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	return color
}

// putBackFaceOntoBattlefield puts a card straight onto the
// battlefield on its BACK face and clears summoning sickness — the
// back-face-ability tests care about the ability, not how the face
// got there (Reflection of Kiki-Jiki's own test takes the same
// shortcut).
func PutBackFaceOntoBattlefield(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := ImportToHand(row, p)
	g.WithWriteLock(func() {
		for i := range p.Hand.Cards {
			if p.Hand.Cards[i].InstanceID == id {
				p.Hand.Cards[i].SetFace(1)
			}
		}
	})
	var entered uuid.UUID
	g.WithWriteLock(func() {
		var err error
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{})
		if err != nil {
			t.Fatalf("put the back face onto the battlefield: %v", err)
		}
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == entered {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return entered
}
