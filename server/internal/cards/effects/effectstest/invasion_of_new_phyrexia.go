package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// invasion_of_new_phyrexia_test.go — S27's last battle (#92, #626),
// both faces.
//
// The front face is an X-drop battle; the back face is the catalog's
// first planeswalker reached by defeating something, and each of its
// three loyalty abilities exercises a different seam: the +1 a
// set-level Validate on the discard pick (#624), the −2 an emblem
// whose text includes a granted WARD (#623 plus WardGranted), the −3
// a CR 603.12 reflexive trigger whose target clause is not knowable
// until the taps are in (#636).

// newPhyrexiaRow is the Scryfall record for Invasion of New Phyrexia
// in the shape deck.ToGameCard consumes: a `transform` layout with a
// null top-level cost, the battle's {X}{W}{U} and defense 6 on face
// 0, and the planeswalker's loyalty on face 1. That is verbatim how
// the card appears in the bulk dump, and driving the test through the
// importer is what proves the back face's LOYALTY survives the trip —
// the same class of bug as the battles' missing defense (#274's twin).
func NewPhyrexiaRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Invasion of New Phyrexia // Teferi Akosa of Zhalfir",
		Layout:   "transform",
		TypeLine: "Battle — Siege // Legendary Planeswalker — Teferi",
		OracleID: uuid.MustParse(effects.InvasionOfNewPhyrexiaOracleID),
		CardFaces: []cards.CardFace{
			{
				Name:     "Invasion of New Phyrexia",
				TypeLine: "Battle — Siege",
				ManaCost: "{X}{W}{U}",
				Defense:  "6",
				Colors:   []string{"W", "U"},
			},
			{
				Name:     "Teferi Akosa of Zhalfir",
				TypeLine: "Legendary Planeswalker — Teferi",
				Loyalty:  "4",
				Colors:   []string{"W", "U"},
			},
		},
	}
}

// pushTeferiAkosa puts the back face on the battlefield directly, for
// the three ability tests that are about the abilities rather than
// about the transformed cast. The catalog key is the composite one
// the back face registers under, which is what CatalogKey yields for
// the real card once its face is turned.
func PushTeferiAkosa(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return PushWalkerForTest(g, owner, "Teferi Akosa of Zhalfir",
		effects.InvasionOfNewPhyrexiaOracleID+"#1", loyalty)
}

func HandIDs(g *game.Game, player uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

func PushToHand(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	p := g.PlayerByIDForEffect(owner)
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// latestChooseCards is the newest choose_cards prompt addressed to a
// player — the -3's "tap any number", here.
func LatestChooseCards(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func InLibrary(g *game.Game, owner, id uuid.UUID) bool {
	p := g.PlayerByIDForEffect(owner)
	if p == nil || p.Library == nil {
		return false
	}
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}
