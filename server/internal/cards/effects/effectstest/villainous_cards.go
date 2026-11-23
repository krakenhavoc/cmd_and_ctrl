package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// villainous_cards_test.go — #3140: the seven cards that use a
// villainous choice and are fully expressible with what the engine has.

const (
	villainThisIsHowItEnds = "53467cf4-5db2-4b9c-af2b-03b044c558cd"
	VillainSycorax         = "78ed3e01-6d22-429a-a9db-baff38d7bc03"
	VillainDavros          = "25c0bbc8-b0c3-4166-967f-e8b096487a6c"
	VillainEggman          = "fb68c1ae-a129-4c17-81fa-be5da2c32ec5"
	VillainGreatPlan       = "cb9abc41-07f8-4a06-b7b9-9dd6be526e42"
)

// villainClearHand empties a player's hand and library of the harness's
// starting cards, so the counts a test makes are its own.
func VillainEmptyHand(p *game.Player) { p.Hand.Cards = nil }

// ---- This Is How It Ends ------------------------------------------------------

func VillainCastEnds(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	CastCatalogSpell(t, g, "This Is How It Ends", "Instant", villainThisIsHowItEnds, CardTarget(target))
	PassPriorityAroundTable(t, g)
}

// ---- Great Intelligence's Plan ---------------------------------------------------

func VillainCastPlan(t *testing.T, g *game.Game, target *game.Player) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Great Intelligence's Plan", TypeLine: "Sorcery", OracleID: VillainGreatPlan,
		Owner: active.ID, Controller: active.ID,
	})
	AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	PassPriorityAroundTable(t, g)
}
