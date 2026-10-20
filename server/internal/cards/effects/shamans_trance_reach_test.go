package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shamans_trance_reach_test.go — #4177: "as though those cards were in
// your graveyard". The caster's own permissions over their graveyard
// (Underworld Breach's escape) reach the cards in every graveyard this
// turn, and nobody else's do.

func trancePushBreach(g *game.Game, controller uuid.UUID) {
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Underworld Breach", TypeLine: "Enchantment",
		OracleID: breachOracle, Owner: controller, Controller: controller,
	})
	g.BumpLayerVersionForTest()
}

func escapeOf(g *game.Game, player, card uuid.UUID) *game.CastPermission {
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		perm = g.CastPermissionForClaimLocked(player, c, game.ZoneGraveyard, "escape")
	})
	if perm != nil && perm.AltCostKey != "escape" {
		return nil
	}
	return perm
}

// Underworld Breach gives the caster's graveyard escape; under the
// Trance the cards in the other graveyards have it too, a land still
// does not, and it can actually be cast that way: the {R} escape cost,
// three of the caster's own other cards exiled.
func TestShamansTranceLetsBreachGiveEscapeToOtherGraveyards(t *testing.T) {
	tt := newTranceTable(t)
	g, me := tt.g, tt.me
	trancePushBreach(g, me.ID)
	if escapeOf(g, me.ID, tt.oppBolt) == nil || escapeOf(g, me.ID, tt.oppLooting) == nil {
		t.Fatal("Underworld Breach gave no escape to a spell in another player's graveyard under the Trance")
	}
	if escapeOf(g, me.ID, tt.oppLand) != nil {
		t.Error("a land gained escape: Breach is \"each nonland card\"")
	}
	var fodder []uuid.UUID
	for _, name := range []string{"Fodder A", "Fodder B", "Fodder C"} {
		fodder = append(fodder, tranceGraveCard(me, name, "Creature — Bear", "", "{1}{G}"))
	}
	life := tt.third.Life
	b06AddMana(me, "R")
	if err := g.CastSpell(me.ID, tt.oppBolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: fodder,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: tt.third.ID}},
		Strict:  true,
	}); err != nil {
		t.Fatalf("escape seat 1's Lightning Bolt with Breach under the Trance: %v", err)
	}
	if !g.Stack.Contains(tt.oppBolt) {
		t.Fatal("the Bolt is not on the stack")
	}
	for _, id := range fodder {
		if !g.Exile.Contains(id) {
			t.Errorf("escape's cost did not exile the caster's card %s", id)
		}
	}
	passPriorityAroundTable(t, g)
	if tt.third.Life != life-3 {
		t.Errorf("seat 2 life = %d, want %d", tt.third.Life, life-3)
	}
}

// The reach is the caster's alone: another player's Breach does not
// reach a third player's graveyard, and the caster's own card keeps its
// escape as it always did.
func TestShamansTranceReachIsTheCastersOnly(t *testing.T) {
	tt := newTranceTable(t)
	g := tt.g
	thirdBolt := tranceGraveCard(tt.third, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}")
	trancePushBreach(g, tt.opp.ID)
	if escapeOf(g, tt.opp.ID, thirdBolt) != nil {
		t.Error("another player's Breach reached a third player's graveyard")
	}
	myBolt := tranceGraveCard(tt.me, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}")
	trancePushBreach(g, tt.me.ID)
	if escapeOf(g, tt.me.ID, myBolt) == nil {
		t.Error("the caster's Breach no longer reaches their own graveyard")
	}
}

// The reach ends with the turn, with the rest of the Trance.
func TestShamansTranceReachEndsWithTheTurn(t *testing.T) {
	tt := newTranceTable(t)
	g, me := tt.g, tt.me
	trancePushBreach(g, me.ID)
	if escapeOf(g, me.ID, tt.oppBolt) == nil {
		t.Fatal("no escape under the Trance")
	}
	passTurnsTo(t, g, 1)
	passTurnsTo(t, g, 0)
	advanceToMain(t, g)
	if escapeOf(g, me.ID, tt.oppBolt) != nil {
		t.Error("the reach outlived the turn")
	}
}

// Restoring the game mid-turn keeps the reach (the mod kind is stored).
func TestShamansTranceReachSurvivesARestore(t *testing.T) {
	tt := newTranceTable(t)
	trancePushBreach(tt.g, tt.me.ID)
	g := restoreRoundTrip(t, tt.g, true)
	if escapeOf(g, g.Seats[0].ID, tt.oppBolt) == nil {
		t.Error("after a restore the caster's Breach no longer reaches seat 1's graveyard")
	}
}
