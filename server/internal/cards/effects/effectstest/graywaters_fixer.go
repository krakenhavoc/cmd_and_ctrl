package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graywaters_fixer.go — Graywater's Fixer test helpers shared across test packages.

// GfDead puts an uncatalogued card into p's graveyard by a real move (a
// discard's zone change), so the zone grant sees it arrive.
func GfDead(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, id); err != nil {
		t.Fatalf("move %s to the graveyard: %v", name, err)
	}
	return id
}

// GfEncoreCost is the mana of the one activated row the card has, or
// ok=false when it has none.
func GfEncoreCost(g *game.Game, id uuid.UUID) (cost string, ok bool) {
	g.ReadSnapshot(func() {
		c, found := g.LookupCardForEffect(id)
		if !found {
			return
		}
		rows := game.ActivatedAbilitiesForCard(c)
		if len(rows) == 1 {
			cost, ok = rows[0].Cost.Mana, true
		}
	})
	return cost, ok
}
