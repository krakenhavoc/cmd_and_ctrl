package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_batch_cards_test.go — the proof cards for #1322, #1324 and
// #1327 (tracker #1306, "Aang is so flashy"): the entry paths a
// shockland, a pay-life MDFC land back, Sword of Hearth and Home and
// Phelia, Exuberant Shepherd each waited on. The engine half is
// server/internal/game/entry_batch_test.go and
// exile_return_then_test.go.

const (
	ebSwordOfHearthAndHomeOracle = "913e6182-706a-4872-8c8a-e146b0ae0738"
	EbPheliaOracle               = "5d86a59a-ba1f-45f7-b829-dca6f9f3e624"
	ebLoyalWarhoundOracle        = "cc6a83c7-e645-4a53-9550-be79b42cd851"
	EbSinkIntoStuporOracle       = "bcc6eece-75ea-494c-b33a-d4477d504e0b"
)

// --- #1324: Sword of Hearth and Home -----------------------------------

// ebSwordSetup equips the Sword to a bear under seat 0, blinks-to-be
// Loyal Warhound beside it, and gives the opponent exactly one more land
// than seat 0 — so Loyal Warhound's intervening-if ("if an opponent
// controls more lands than you") is TRUE before the Sword's basic land
// arrives and FALSE after.
func EbSwordSetup(t *testing.T) (g *game.Game, me, opp *game.Player, bear, warhound, plains, spare uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	AdvanceToMain(t, g)
	bear = SeedBear(g, me.ID)
	sword := SeedEquipment(g, me.ID, "Sword of Hearth and Home", ebSwordOfHearthAndHomeOracle)
	EquipTo(t, g, me.ID, sword, bear)
	warhound = PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loyal Warhound", TypeLine: "Creature — Dog",
		OracleID: ebLoyalWarhoundOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 2; i++ {
		PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
			Owner: me.ID, Controller: me.ID,
		})
	}
	for i := 0; i < 3; i++ {
		PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp",
			Owner: opp.ID, Controller: opp.ID,
		})
	}
	plains = PlTop(me, "Plains", "Basic Land — Plains", "")
	PlTop(me, "Opt", "Instant", "{U}")
	spare = PlTop(me, "Plains", "Basic Land — Plains", "")
	return g, me, opp, bear, warhound, plains, spare
}
