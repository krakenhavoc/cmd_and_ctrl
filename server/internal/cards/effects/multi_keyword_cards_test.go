package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// multi_keyword_cards_test.go covers the S18 sub-PR 5 batch 2
// catalog cards. Each card carries ≥2 keywords; the tests assert
// every keyword lands in Effective().Abilities on the battlefield,
// AND that off-battlefield HasKeyword returns true via the
// catalog fallback (the forcing function for flash gating on
// hand-resident Ambush Viper).

const (
	vampireNighthawkOracle = "feb244f8-bcb1-44cf-9940-2719221a7309"
	baneslayerAngelOracle  = "0e11792b-7fe5-4208-aa0b-e5d09b2b65fe"
	ambushViperOracle      = "8957f7c2-040c-4048-9f21-efa7c97682b7"
	boggartBruteOracle     = "880793a2-84b0-4ae1-9bb8-6e942e3dc723"
)

func TestVampireNighthawkHasFlyingDeathtouchLifelink(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Vampire Nighthawk",
		TypeLine:   "Creature — Vampire Shaman",
		OracleID:   vampireNighthawkOracle,
		Power:      2,
		Toughness:  3,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "flying", "deathtouch", "lifelink")
}

func TestBaneslayerAngelHasFlyingFirstStrikeLifelink(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Baneslayer Angel",
		TypeLine:   "Creature — Angel",
		OracleID:   baneslayerAngelOracle,
		Power:      5,
		Toughness:  5,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	// Protection clauses NOT expected — scoped out to S24 per ADR 0014.
	assertKeywords(t, g, id, "flying", "first strike", "lifelink")
}

func TestAmbushViperHasFlashDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Ambush Viper",
		TypeLine:   "Creature — Snake",
		OracleID:   ambushViperOracle,
		Power:      2,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "flash", "deathtouch")
}

func TestAmbushViperFlashReadableInHand(t *testing.T) {
	// Flash is the forcing function for the off-battlefield
	// PrintedKeywords fallback — CastSpell's flash gate runs against
	// a hand-resident card.
	c := &game.Card{
		InstanceID: uuid.New(),
		OracleID:   ambushViperOracle,
		TypeLine:   "Creature — Snake",
	}
	if !game.HasKeyword(c, "flash") {
		t.Errorf("hand-resident Ambush Viper: expected flash via catalog fallback")
	}
	if !game.HasKeyword(c, "deathtouch") {
		t.Errorf("hand-resident Ambush Viper: expected deathtouch via catalog fallback")
	}
}

func TestBoggartBruteHasMenace(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Boggart Brute",
		TypeLine:   "Creature — Goblin Warrior",
		OracleID:   boggartBruteOracle,
		Power:      3,
		Toughness:  2,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "menace")
}
