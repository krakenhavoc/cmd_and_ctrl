package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_celebration_cards_test.go — celebration (CR 207.2c, #5164) and the
// Wilds of Eldraine cards that read it. The seam itself (the tally cell)
// is pinned in game/turn_tally_celebration_test.go; these play each
// card's real catalog entry.

const (
	ArmoryMiceOracle            = "68f7d1bf-48c7-483e-b17f-a60008854b46"
	AshPartyCrasherOracle       = "1bc701e2-913a-45d5-bc3f-16f87e7e2653"
	BelligerentOfTheBallOracle  = "2be2f6c8-d8df-4e5c-a0c1-7566183bbe81"
	BespokeBattlegarbOracle     = "b2ff15b8-1be4-4f94-b4cc-aa959a50288e"
	CelebrationSoldierTokenName = "Soldier"
)

// --- fixtures ------------------------------------------------------------------

// celebrationTokens creates n 1/1 Soldier tokens under owner's control:
// n nonland permanents entering.
func CelebrationTokens(t *testing.T, g *game.Game, owner uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(owner, game.Card{Name: CelebrationSoldierTokenName, TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}, n); err != nil {
			t.Fatalf("create tokens: %v", err)
		}
	})
}

// celebrationEnterFromHand puts a card from owner's hand onto the
// battlefield through the ordinary entry pipeline, so its entry counts.
func CelebrationEnterFromHand(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		g.PlayerByIDForEffect(owner).Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: 1, Toughness: 1, Owner: owner, Controller: owner,
		})
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Controller: owner})
	})
	if err != nil || entered == uuid.Nil {
		t.Fatalf("put %s onto the battlefield: %v", name, err)
	}
	return entered
}
