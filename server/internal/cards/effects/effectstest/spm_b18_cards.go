package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b18_cards_test.go — Marvel's Spider-Man, batch 18 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	SpmB18PassageOracle   = "17770eeb-ecbf-4510-af2d-acececfbe146"
	SpmB18MysterioOracle  = "40f664d8-4c73-44aa-8fb1-d38771a42520"
	SpmB18NormanOracle    = "aa5b06a4-90c3-45e7-ab83-cb5c96ded21d"
	SpmB18OriginOracle    = "7fbe9056-190f-40a1-bee9-59ebd6f981f5"
	SpmB18ParkerOracle    = "03d1a4e9-3ebc-491f-8741-5d933b500abf"
	SpmB18SandmanOracle   = "8b205480-b1fb-490c-b1c9-cf5cc3c5acdb"
	SpmB18ScarletOracle   = "13ab2f3f-f65e-42b5-b743-39b30a4c4751"
	SpmB18VerseOracle     = "7ef7d902-1f05-40b3-8997-dec197b0b678"
	SpmB18WomanOracle     = "be2b9c6d-4ecb-49ec-b276-4aa93c5dfc00"
	SpmB18SymbioteOracle  = "bc681ecf-2a52-4782-9a47-90f5a9f435d8"
	SpmB18SymbioteType    = "Legendary Creature — Symbiote Spider Hero"
	SpmB18SandmanTypeLine = "Legendary Creature — Sand Elemental Villain"
)

// spmB18Graveyard puts a card into p's graveyard.
func SpmB18Graveyard(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

// --- Norman Osborn // Green Goblin ---

func SpmB18NormanCard(owner uuid.UUID, face int) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: SpmB18NormanOracle, Layout: game.LayoutModalDFC,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Norman Osborn", TypeLine: "Legendary Creature — Human Scientist Villain", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 1, Toughness: 1},
			{Name: "Green Goblin", TypeLine: "Legendary Creature — Goblin Human Villain", ManaCost: "{1}{U}{B}{R}", Colors: []string{"U", "B", "R"}, Power: 3, Toughness: 3},
		},
	}
	c.SetFace(face)
	return c
}
