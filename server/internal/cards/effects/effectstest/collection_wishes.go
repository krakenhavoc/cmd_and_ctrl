package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// collection_wishes_test.go — ADR 0198: a deck with no sideboard takes "a
// card you own from outside the game" from the card database. Every card
// built on TakeFromOutside gains that without a change to its file; these
// pin it card by card, and that the card taken is a real catalog card.

const CultivateOracle = "8b755881-a72d-4e21-a369-d2924eb4585a"

type TestCollection struct{ Cards []game.Card }

func (f TestCollection) Search(q string, limit int) []game.Card {
	var out []game.Card
	for _, c := range f.Cards {
		if strings.Contains(strings.ToLower(c.Name), strings.ToLower(q)) && len(out) < limit {
			out = append(out, c)
		}
	}
	return out
}

func (f TestCollection) Named(name string) (game.Card, bool) {
	for _, c := range f.Cards {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return game.Card{}, false
}

func (f TestCollection) Suggest(fn func(game.Card) bool) {
	for _, c := range f.Cards {
		if !fn(c) {
			return
		}
	}
}

// dbCard is a card-database template.
func DbCard(name, typeLine, oracle string, colors ...string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine = typeLine
	c.OracleID = oracle
	c.Colors = colors
	c.ScryfallID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(name)).String()
	return c
}

// installTestCollection installs a database with one card of each kind a
// wish asks for.
func InstallTestCollection(t *testing.T) {
	t.Helper()
	prev := game.OutsideCollection
	game.OutsideCollection = TestCollection{Cards: []game.Card{
		DbCard("Lightning Bolt", "Instant", LightningBoltOracle, "R"),
		DbCard("Cultivate", "Sorcery", CultivateOracle, "G"),
		DbCard("Sol Ring", "Artifact", SolRingOracle),
		DbCard("Grizzly Bears", "Creature — Bear", "", "G"),
		DbCard("Forest", "Basic Land — Forest", ""),
		DbCard("Pacifism", "Enchantment — Aura", "", "W"),
		DbCard("Lightning Helix", "Instant", "", "R", "W"),
		DbCard("Ulamog's Crusher", "Creature — Eldrazi", ""),
	}}
	t.Cleanup(func() { game.OutsideCollection = prev })
}

func LatestCollectionPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceOutsideCollection && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func HandCardNamed(p *game.Player, name string) (game.Card, bool) {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}
