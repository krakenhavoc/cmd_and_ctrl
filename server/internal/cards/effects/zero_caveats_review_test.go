package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zero_caveats_review_test.go — #4177: tests for catalog cards the
// Completeness review found with none, and for the one divergence it
// fixed (Pawn of Ulamog's token copy).

// Palladium Myr: "{T}: Add {C}{C}." An artifact CREATURE, so summoning
// sickness applies (CR 302.6) the turn it comes in.
func TestPalladiumMyrTapsForTwoColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	myr := pushCatalogPermanent(g, me.ID, "Palladium Myr", "Artifact Creature — Myr", palladiumMyrOracle, false)
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Palladium Myr: %v", err)
	}
	colorless := 0
	for _, c := range batch01PoolColors(me) {
		if c == "C" {
			colorless++
		}
	}
	if colorless != 2 {
		t.Errorf("pool has %d colorless, want 2", colorless)
	}
	if card, ok := g.LookupCardForEffect(myr); !ok || !card.Tapped {
		t.Error("the Myr is not tapped after its ability")
	}
}

func TestPalladiumMyrIsSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	myr := pushCatalogPermanent(g, me.ID, "Palladium Myr", "Artifact Creature — Myr", palladiumMyrOracle, true)
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrSummoningSick) {
		t.Errorf("a summoning-sick Myr tapped for mana: err = %v", err)
	}
}

// Pawn of Ulamog: "Whenever this creature or another nontoken creature
// you control dies". A TOKEN copy of the Pawn dying is still "this
// creature"; any other token is not.
func TestPawnOfUlamogTokenCopyOfItselfStillTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Pawn of Ulamog", TypeLine: "Token Creature — Vampire Shaman",
		OracleID: pawnOfUlamogOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	g.BumpLayerVersionForTest()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Eldrazi Spawn"); got != 1 {
		t.Errorf("a token Pawn dying made %d Spawn, want 1", got)
	}
}
