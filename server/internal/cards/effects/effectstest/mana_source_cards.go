package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_source_cards_test.go — the catalog half of #1212: three cards
// that read the payment that made them.
//
// The engine-side contract is pinned in game/mana_source_test.go. What
// is pinned here is that the three printed shapes reach it: a
// permanent asking about the SOURCE of its mana (Hired Hexblade), a
// permanent asking about a COLOUR (Gruul Scrapper), and a spell asking
// about a colour while it resolves (Ribbons of Night).

const (
	HiredHexbladeOracle = "f3a0f155-05d8-465c-b0ef-35aa12e93013"
	GruulScrapperOracle = "878e8980-a2cc-4335-aafe-f5166ef48f79"
	RibbonsOfNightOracl = "f074c0d0-8455-484d-ae75-820fdfbc4740"
)

// crackATreasureFor spawns a real Treasure token, cracks it, and
// answers its colour pick — leaving one token of `color` in the pool
// that remembers it came from a Treasure.
func CrackATreasureFor(t *testing.T, g *game.Game, p *game.Player, color string) {
	t.Helper()
	tmpl, ok := effects.Tokens().TokenTemplate("Treasure")
	if !ok {
		t.Fatal("no Treasure template")
	}
	ids, err := g.SpawnCards(p.ID, p.ID, game.ZoneBattlefield, tmpl, 1)
	if err != nil || len(ids) != 1 {
		t.Fatalf("SpawnCards: %v (%d ids)", err, len(ids))
	}
	if err := g.ActivateManaAbility(p.ID, ids[0], 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack the Treasure: %v", err)
	}
	pick := RiderLatestManaPick(g, p.ID)
	if pick == nil {
		t.Fatal("no colour pick queued by the Treasure")
	}
	if err := g.ResolveManaChoice(pick.ID, p.ID, color); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
}

// castFromHandForTest puts a card in hand and casts it under `params`.
func CastFromHandForTest(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}
