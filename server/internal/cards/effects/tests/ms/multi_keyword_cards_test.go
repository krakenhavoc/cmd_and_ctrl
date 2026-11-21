package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestVampireNighthawkHasFlyingDeathtouchLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Vampire Nighthawk",
		TypeLine:   "Creature — Vampire Shaman",
		OracleID:   effectstest.VampireNighthawkOracle,
		Power:      2,
		Toughness:  3,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "flying", "deathtouch", "lifelink")
}

func TestBaneslayerAngelHasFlyingFirstStrikeLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Baneslayer Angel",
		TypeLine:   "Creature — Angel",
		OracleID:   effectstest.BaneslayerAngelOracle,
		Power:      5,
		Toughness:  5,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	// Protection clauses NOT expected — scoped out to S24 per ADR 0014.
	effectstest.AssertKeywords(t, g, id, "flying", "first strike", "lifelink")
}

func TestAmbushViperHasFlashDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Ambush Viper",
		TypeLine:   "Creature — Snake",
		OracleID:   effectstest.AmbushViperOracle,
		Power:      2,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "flash", "deathtouch")
}

func TestAmbushViperFlashReadableInHand(t *testing.T) {
	// Flash is the forcing function for the off-battlefield
	// PrintedKeywords fallback — CastSpell's flash gate runs against
	// a hand-resident card.
	c := &game.Card{
		InstanceID: uuid.New(),
		OracleID:   effectstest.AmbushViperOracle,
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
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Boggart Brute",
		TypeLine:   "Creature — Goblin Warrior",
		OracleID:   effectstest.BoggartBruteOracle,
		Power:      3,
		Toughness:  2,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "menace")
}
