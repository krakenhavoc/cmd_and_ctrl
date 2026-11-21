package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_r37_helpers_test.go — test helpers the S82 GAP batch 100–103
// tests borrowed from other test files. Since the #5906 split those files
// sit in other test packages, so roll-up 37 gives this package its own
// copies.

// r37Sliver seeds a non-catalog Sliver creature that has been under its
// controller's control since the turn began (tspB1Sliver's shape).
func r37Sliver(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Sliver",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// r37Dead puts an uncatalogued card into p's graveyard by a real move (a
// discard's zone change), so a zone grant sees it arrive (gfDead's shape).
func r37Dead(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
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

// r37EncoreCost is the mana of the one activated row the card has, or
// ok=false when it has none (gfEncoreCost's shape).
func r37EncoreCost(g *game.Game, id uuid.UUID) (cost string, ok bool) {
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

// r37Tokens counts the tokens named name that controller controls.
func r37Tokens(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller && c.IsToken() {
				n++
			}
		}
	})
	return n
}

// r37Benalia puts a History of Benalia (three chapters) onto the
// battlefield under p with lore lore counters already on it.
func r37Benalia(g *game.Game, p *game.Player, lore int) uuid.UUID {
	return effectstest.ApaPush(g, p.ID, p.ID, game.Card{Name: "History of Benalia", OracleID: effectstest.HistoryOfBenaliaOracle,
		TypeLine: "Enchantment — Saga", ManaCost: "{1}{W}{W}", Counters: map[string]int{game.CounterLore: lore}})
}

// r37Advance puts one lore counter on the Saga and lets everything that
// triggers resolve.
func r37Advance(t *testing.T, g *game.Game, saga uuid.UUID) {
	t.Helper()
	if err := g.AddCounter(saga, game.CounterLore, 1); err != nil {
		t.Fatalf("add a lore counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// r37Toughness is the card's current toughness after the layers are fresh.
func r37Toughness(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("no card %s", id)
		}
		n = c.CurrentToughness()
	})
	return n
}

// r37GivePoison puts n poison counters on p.
func r37GivePoison(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(p.ID, game.CounterPoison, n) })
}
