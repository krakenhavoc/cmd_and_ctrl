package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// play_from_outside_test.go — #3955: Spawnsire of Ulamog's {20} and Wish,
// over both pools (a deck's sideboard, the card database).

const (
	WishOracle           = "2ccbf0ee-c856-4f67-8579-811fe8571a2e"
	templeOfAbandonOracl = "3baa8e38-ef93-435d-b63e-f781d5bfcc68"
)

// fundOutside puts `n` colorless mana in the pool, for a paid play.
func FundOutside(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}

func CastFromOutside(g *game.Game, p *game.Player, id uuid.UUID, strict bool) error {
	return g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "outside", Strict: strict})
}

// installOutsideCollection is the test database for the cast tests.
func InstallOutsideCollection(t *testing.T) {
	t.Helper()
	prev := game.OutsideCollection
	game.OutsideCollection = TestCollection{Cards: []game.Card{
		DbCard("Grizzly Bears", "Creature — Bear", "", "G"),
		DbCard("Forest", "Basic Land — Forest", ""),
		DbCard("Temple of Abandon", "Land", templeOfAbandonOracl),
		DbCard("Steam Vents", "Land — Island Mountain", SteamVentsOracle),
		DbCard("Ulamog's Crusher", "Creature — Eldrazi", "", ""),
		DbCard("Emrakul's Hatcher", "Creature — Eldrazi Drone", "", ""),
		DbCard("Lightning Bolt", "Instant", LightningBoltOracle, "R"),
	}}
	t.Cleanup(func() { game.OutsideCollection = prev })
}

func WithManaCost(c game.Card, cost string) game.Card {
	c.ManaCost = cost
	return c
}

// --- Spawnsire of Ulamog ------------------------------------------------

func SpawnsireReady(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	ToMain(t, g)
	return PushAnnCatalogCreature(g, me.ID, "Spawnsire of Ulamog", SpawnsireOracle, 7, 11)
}
