package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// empower_jace_test.go — ADR 0139 (#2796): the Jace planeswalker token
// and the Empower Jace keyword action (CR 701.71a), then one test per
// proof card.

const (
	OracleNoAdmittance          = "d53ae01b-87e1-48e1-9fc3-23a734ce6622"
	OracleCountersculpt         = "fc3353e9-fe24-4ca7-bff8-9da767f2903a"
	OracleRewriteRegrets        = "42565432-3a17-4274-87fc-016ab900e313"
	OracleCampusCrier           = "c1a9d60f-a469-48f9-8f1c-3ca5ae86bd64"
	OracleSolveForDisappoint    = "3241a3f5-7ae5-4497-9dd2-a55f3370a055"
	OraclePlanForAllOutcomes    = "7fd2cac5-cd29-4a47-9e89-9c6be39a0b67"
	OracleMindseekerOculus      = "aa9547b3-e4e4-4557-9dd1-3aa9a7cb9937"
	OracleJaceRealitySculptor   = "bf0e9e00-fba5-448a-9052-ae730aea4bf8"
	OracleAcademicAscent        = "406e853b-713f-4c89-9efb-aca74a11183f"
	oracleProtegesAwakening     = "07617648-6490-4719-acec-7705671f7e01"
	OracleTamsResistance        = "7500b7dc-588d-46bd-acc5-885dcd464668"
	OracleVraskasFinalMercy     = "7799413d-e313-4a05-ad28-f4c926c26d28"
	OracleArcaneAmphisbaena     = "5c151e3b-fcf2-4adb-a60b-9eff0c360705"
	OracleKeeperOfTheQuietHour  = "ecdfee75-15fd-41ea-89d4-4bfc1fa4d032"
	OracleRepurposedEnforcer    = "b775a404-ec0a-47ff-bb9d-eae2273d945b"
	JaceRealitySculptorTypeLine = "Legendary Planeswalker — Jace"
)

// jaceTokensOf lists the Jace planeswalker tokens a player controls.
func JaceTokensOf(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.WithWriteLock(func() { out = effects.JaceTokensControlledBy(g, player) })
	return out
}

// onlyJaceToken returns the one Jace token a player controls, failing
// the test when there is not exactly one.
func OnlyJaceToken(t *testing.T, g *game.Game, player uuid.UUID) uuid.UUID {
	t.Helper()
	jaces := JaceTokensOf(g, player)
	if len(jaces) != 1 {
		t.Fatalf("player controls %d Jace tokens, want 1", len(jaces))
	}
	return jaces[0]
}

// seedJaceToken creates a Jace token for player through the ordinary
// creation path and puts `loyalty` counters on it, in one locked call
// so no state-based action check sees it empty.
func SeedJaceToken(t *testing.T, g *game.Game, player uuid.UUID, loyalty int) uuid.UUID {
	t.Helper()
	var made []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		made, err = g.CreateTokensForEffect(player, effects.JaceToken(), 1, game.TokenEntryOptions{})
		if err == nil && len(made) == 1 {
			err = g.AddCounterForEffect(made[0], game.CounterLoyalty, loyalty)
		}
	})
	if err != nil || len(made) != 1 {
		t.Fatalf("seed a Jace token: made %v, err %v", made, err)
	}
	return made[0]
}

// castProtegesAwakening casts Protege's Awakening (Empower Jace 6, then
// draw a card), the plainest card on the seam.
func CastProtegesAwakening(t *testing.T, g *game.Game) {
	t.Helper()
	CastCatalogSpell(t, g, "Protege's Awakening", "Sorcery", oracleProtegesAwakening, nil)
}
